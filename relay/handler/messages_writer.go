package controller

// anthropicResponseWriter wraps gin.ResponseWriter and translates the OpenAI
// chat protocol flowing out of RelayTextHelper into the Anthropic Messages
// protocol:
//
//   - non-stream: the OpenAI JSON body is buffered and converted once at
//     Finalize; error envelopes become Anthropic error envelopes;
//   - stream: each `data: {chunk}` SSE line is translated on the fly into
//     message_start / content_block_start / content_block_delta /
//     content_block_stop / message_delta / message_stop.
//
// It deliberately sits ON TOP of the RelayState writer so failover state and
// content capture still see exactly the bytes the client received.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	relaymodel "github.com/perryles/one-api-next/relay/schema"
)

type anthropicResponseWriter struct {
	gin.ResponseWriter
	underlying    gin.ResponseWriter
	originModel   string
	stream        bool
	stopSequences []string
	inputEstimate int

	// stream state
	started      bool
	lineBuf      bytes.Buffer
	textOpen     bool        // a text content block is currently open
	nextIndex    int         // next Anthropic content block index
	toolIndex    map[int]int // openai tool_calls index → anthropic block index
	toolOpened   map[int]bool
	fullText     strings.Builder
	finishReason string
	outputTokens int
	usageSeen    bool
	ended        bool

	// non-stream state
	status  int
	bodyBuf bytes.Buffer
}

func newAnthropicResponseWriter(w gin.ResponseWriter, originModel string, stream bool, stopSequences []string, inputEstimate int) *anthropicResponseWriter {
	return &anthropicResponseWriter{
		ResponseWriter: w,
		underlying:     w,
		originModel:    originModel,
		stream:         stream,
		stopSequences:  stopSequences,
		inputEstimate:  inputEstimate,
		toolIndex:      map[int]int{},
		toolOpened:     map[int]bool{},
	}
}

// ---- gin.ResponseWriter plumbing ----

func (w *anthropicResponseWriter) WriteHeader(code int) {
	w.status = code
	// hold the header back until Finalize: the body length changes
	// during conversion, so any upstream Content-Length must not leak
}

func (w *anthropicResponseWriter) WriteHeaderNow() {}

func (w *anthropicResponseWriter) Write(data []byte) (int, error) {
	if w.stream {
		w.feedStream(data)
	} else {
		w.bodyBuf.Write(data)
	}
	return len(data), nil
}

func (w *anthropicResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *anthropicResponseWriter) Flush() {
	if w.stream {
		// SSE events are already written straight through emit()
		w.underlying.Flush()
	}
}

func (w *anthropicResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

// ---- stream translation ----

func (w *anthropicResponseWriter) feedStream(data []byte) {
	w.lineBuf.Write(data)
	for {
		line, err := w.lineBuf.ReadString('\n')
		if err != nil {
			// incomplete line: push back the remainder
			w.lineBuf = *bytes.NewBufferString(line)
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data: ") && !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " ")
		w.consumeSSE(payload)
	}
}

func (w *anthropicResponseWriter) consumeSSE(payload string) {
	if payload == "[DONE]" {
		w.endStream()
		return
	}
	var chunk struct {
		Id      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Index int `json:"index"`
			Delta struct {
				Role      string            `json:"role"`
				Content   any               `json:"content"`
				ToolCalls []json.RawMessage `json:"tool_calls"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage *relaymodel.Usage `json:"usage"`
	}
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return // upstream noise — never forward malformed chunks
	}
	if !w.started {
		w.started = true
		w.emitJSON("message_start", map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id":            "msg-" + strings.TrimPrefix(chunk.Id, "chatcmpl-"),
				"type":          "message",
				"role":          "assistant",
				"model":         w.originModel,
				"content":       []any{},
				"stop_reason":   nil,
				"stop_sequence": nil,
				"usage": map[string]any{
					"input_tokens":  w.inputEstimate,
					"output_tokens": 0,
				},
			},
		})
		w.emitJSON("ping", map[string]any{"type": "ping"})
	}
	if chunk.Usage != nil {
		w.usageSeen = true
		if chunk.Usage.CompletionTokens > 0 {
			w.outputTokens = chunk.Usage.CompletionTokens
		}
	}
	for _, choice := range chunk.Choices {
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			w.finishReason = *choice.FinishReason
			continue
		}
		// text deltas
		if text := deltaText(choice.Delta.Content); text != "" {
			if !w.textOpen {
				w.textOpen = true
				w.emitJSON("content_block_start", map[string]any{
					"type":          "content_block_start",
					"index":         w.nextIndex,
					"content_block": map[string]any{"type": "text", "text": ""},
				})
				w.nextIndex++
			}
			w.fullText.WriteString(text)
			w.emitJSON("content_block_delta", map[string]any{
				"type":  "content_block_delta",
				"index": w.nextIndex - 1,
				"delta": map[string]any{"type": "text_delta", "text": text},
			})
		}
		// tool_call deltas
		for _, raw := range choice.Delta.ToolCalls {
			var tc struct {
				Index    int    `json:"index"`
				Id       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			}
			if err := json.Unmarshal(raw, &tc); err != nil {
				continue
			}
			if !w.toolOpened[tc.Index] {
				// a tool call starts a new block; close the text block first
				if w.textOpen {
					w.emitJSON("content_block_stop", map[string]any{
						"type":  "content_block_stop",
						"index": w.nextIndex - 1,
					})
					w.textOpen = false
				}
				w.toolOpened[tc.Index] = true
				w.toolIndex[tc.Index] = w.nextIndex
				w.nextIndex++
				w.emitJSON("content_block_start", map[string]any{
					"type":  "content_block_start",
					"index": w.toolIndex[tc.Index],
					"content_block": map[string]any{
						"type":  "tool_use",
						"id":    tc.Id,
						"name":  tc.Function.Name,
						"input": map[string]any{},
					},
				})
			}
			if tc.Function.Arguments != "" {
				w.emitJSON("content_block_delta", map[string]any{
					"type":  "content_block_delta",
					"index": w.toolIndex[tc.Index],
					"delta": map[string]any{"type": "input_json_delta", "partial_json": tc.Function.Arguments},
				})
			}
		}
	}
}

func deltaText(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		var sb strings.Builder
		for _, part := range v {
			if m, ok := part.(map[string]any); ok {
				if t, _ := m["type"].(string); t == "text" {
					if text, _ := m["text"].(string); text != "" {
						sb.WriteString(text)
					}
				}
			}
		}
		return sb.String()
	}
	return ""
}

func (w *anthropicResponseWriter) endStream() {
	if w.ended {
		return
	}
	w.ended = true
	if w.textOpen {
		w.emitJSON("content_block_stop", map[string]any{"type": "content_block_stop", "index": w.nextIndex - 1})
		w.textOpen = false
	}
	for _, idx := range w.toolIndex {
		w.emitJSON("content_block_stop", map[string]any{"type": "content_block_stop", "index": idx})
	}
	stopReason := finishToStopReason(w.finishReason)
	var stopSequence any
	if stopReason == "end_turn" {
		if seq, ok := stopSequenceHit(w.fullText.String(), w.stopSequences); ok {
			stopReason = "stop_sequence"
			stopSequence = seq
		}
	}
	w.emitJSON("message_delta", map[string]any{
		"type":  "message_delta",
		"delta": map[string]any{"stop_reason": stopReason, "stop_sequence": stopSequence},
		"usage": map[string]any{"output_tokens": w.outputTokens},
	})
	w.emitJSON("message_stop", map[string]any{"type": "message_stop"})
}

func (w *anthropicResponseWriter) emitJSON(event string, payload map[string]any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	eventLine := "event: " + event + "\n"
	dataLine := "data: " + string(data) + "\n\n"
	_, _ = w.underlying.Write([]byte(eventLine))
	_, _ = w.underlying.Write([]byte(dataLine))
}

// ---- finalize ----

// Finalize flushes everything that was held back. failed marks that the relay
// returned an error (the buffered body is then an OpenAI error envelope that
// still needs conversion — the status may be an error code).
func (w *anthropicResponseWriter) Finalize(failed bool) {
	if failed && !w.started && w.bodyBuf.Len() == 0 {
		return
	}
	if w.stream && (failed || (!w.ended && w.finishReason == "")) {
		w.emitJSON("error", map[string]any(anthropicErrorJSON("api_error", "upstream stream failed")))
		w.ended = true
		w.underlying.Flush()
		return
	}
	if w.stream {
		if !w.ended {
			w.endStream()
		}
		w.underlying.Flush()
		return
	}
	body := w.bodyBuf.Bytes()
	status := w.Status()
	header := w.underlying.Header()
	// error envelope (any status or OpenAI {"error":...} body) → Anthropic
	if converted := openAIErrorBodyToAnthropic(body); converted != nil {
		header.Del("Content-Length")
		header.Set("Content-Type", "application/json")
		w.underlying.WriteHeader(status)
		_, _ = w.underlying.Write(converted)
		return
	}
	converted := convertOpenAIChatToAnthropic(body, w.originModel, w.stopSequences, w.inputEstimate)
	if converted == nil {
		converted = body // unparseable — forward verbatim rather than eat it
	}
	header.Del("Content-Length")
	w.underlying.WriteHeader(status)
	_, _ = w.underlying.Write(converted)
}
