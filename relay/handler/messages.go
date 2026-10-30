package controller

// Anthropic Messages client protocol (POST /v1/messages, /v1/messages/count_tokens).
//
// Design: boundary conversion. The Anthropic request is converted to OpenAI
// chat format and pushed through the standard text relay pipeline
// (RelayTextHelper) so existing billing, subscription and routing apply, while a
// wrapping ResponseWriter translates the OpenAI JSON / SSE stream back into
// Anthropic Messages format (message_start / content_block_* / message_delta /
// message_stop) on the fly.
//
// This makes Claude Code and every Anthropic-SDK client work against ALL
// upstream channel types (OpenAI-compatible, native Anthropic channels via
// their adaptor, etc.), not just Anthropic upstreams.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/perryles/one-api-next/common"
	"github.com/perryles/one-api-next/common/ctxkey"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	relaymodel "github.com/perryles/one-api-next/relay/schema"
)

// ------------------------------------------------------------ request ----

// anthropicContentBlock is one block of an Anthropic message content array.
type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// tool_use
	Id    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
	// tool_result
	ToolUseId string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
	// image
	Source *anthropicImageSource `json:"source,omitempty"`
}

type anthropicImageSource struct {
	Type      string `json:"type"` // base64 | url
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	Url       string `json:"url,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content json.RawMessage         `json:"content"` // string | []block
	Blocks  []anthropicContentBlock `json:"-"`       // filled by parse
	WasText bool                    `json:"-"`
}

type anthropicTool struct {
	Type        string          `json:"type,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema,omitempty"`
}

type anthropicToolChoice struct {
	Type string `json:"type"` // auto | any | tool
	Name string `json:"name,omitempty"`
}

// anthropicRequest is the Anthropic Messages API body.
type anthropicRequest struct {
	Model         string               `json:"model"`
	Messages      []anthropicMessage   `json:"messages"`
	System        json.RawMessage      `json:"system,omitempty"` // string | []{type:text,text}
	MaxTokens     int                  `json:"max_tokens"`
	Stream        bool                 `json:"stream"`
	Temperature   *float64             `json:"temperature,omitempty"`
	TopP          *float64             `json:"top_p,omitempty"`
	TopK          *int                 `json:"top_k,omitempty"`
	StopSequences []string             `json:"stop_sequences,omitempty"`
	Tools         []anthropicTool      `json:"tools,omitempty"`
	ToolChoice    *anthropicToolChoice `json:"tool_choice,omitempty"`
	Metadata      json.RawMessage      `json:"metadata,omitempty"`
}

func (m *anthropicMessage) parseContent() {
	if len(m.Content) == 0 {
		return
	}
	if m.Content[0] == '"' {
		var s string
		if err := json.Unmarshal(m.Content, &s); err == nil {
			m.WasText = true
			m.Blocks = []anthropicContentBlock{{Type: "text", Text: s}}
		}
		return
	}
	var blocks []anthropicContentBlock
	if err := json.Unmarshal(m.Content, &blocks); err == nil {
		m.Blocks = blocks
	}
}

// anthropicToOpenAI converts the request. Returns the OpenAI chat request
// plus the pieces needed on the way back (stop sequences, origin model).
func anthropicToOpenAI(req *anthropicRequest) (*relaymodel.GeneralOpenAIRequest, error) {
	out := &relaymodel.GeneralOpenAIRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}
	if out.MaxTokens <= 0 {
		out.MaxTokens = 4096
	}
	if len(req.StopSequences) > 0 {
		out.Stop = req.StopSequences
	}
	if req.Stream {
		// usage accounting for streaming bills by the final usage chunk
		out.StreamOptions = &relaymodel.StreamOptions{IncludeUsage: true}
	}
	// system: string or text blocks
	if sys := anthropicTextOf(req.System); sys != "" {
		out.Messages = append(out.Messages, relaymodel.Message{Role: "system", Content: sys})
	}
	for i := range req.Messages {
		msg := &req.Messages[i]
		msg.parseContent()
		switch {
		case msg.Role == "user" && hasToolResult(msg.Blocks):
			// a user turn carrying tool_result(s) becomes one or more
			// role:tool messages; interleaved text becomes a user message
			var texts []string
			for _, b := range msg.Blocks {
				if b.Type == "tool_result" {
					content := anthropicTextOf(b.Content)
					if content == "" && len(b.Content) > 0 {
						content = string(b.Content)
					}
					out.Messages = append(out.Messages, relaymodel.Message{
						Role:       "tool",
						ToolCallId: b.ToolUseId,
						Content:    content,
					})
					continue
				}
				if b.Type == "text" {
					texts = append(texts, b.Text)
				}
			}
			if len(texts) > 0 {
				out.Messages = append(out.Messages, relaymodel.Message{
					Role:    "user",
					Content: strings.Join(texts, "\n"),
				})
			}
		case msg.Role == "assistant":
			var texts []string
			var toolCalls []relaymodel.Tool
			for _, b := range msg.Blocks {
				switch b.Type {
				case "text":
					texts = append(texts, b.Text)
				case "tool_use":
					args := b.Input
					if len(args) == 0 {
						args = json.RawMessage("{}")
					}
					toolCalls = append(toolCalls, relaymodel.Tool{
						Id:   b.Id,
						Type: "function",
						Function: relaymodel.Function{
							Name:      b.Name,
							Arguments: string(args),
						},
					})
				}
			}
			m := relaymodel.Message{Role: "assistant"}
			if len(texts) > 0 {
				m.Content = strings.Join(texts, "\n")
			}
			m.ToolCalls = toolCalls
			out.Messages = append(out.Messages, m)
		default:
			// plain user turn (text / image blocks)
			if len(msg.Blocks) == 1 && msg.Blocks[0].Type == "text" {
				out.Messages = append(out.Messages, relaymodel.Message{
					Role:    msg.Role,
					Content: msg.Blocks[0].Text,
				})
				continue
			}
			parts := make([]relaymodel.MessageContent, 0, len(msg.Blocks))
			for _, b := range msg.Blocks {
				switch b.Type {
				case "text":
					parts = append(parts, relaymodel.MessageContent{
						Type: "text",
						Text: b.Text,
					})
				case "image":
					url := anthropicImageURL(b.Source)
					if url != "" {
						parts = append(parts, relaymodel.MessageContent{
							Type:     "image_url",
							ImageURL: &relaymodel.ImageURL{Url: url},
						})
					}
				}
			}
			if len(parts) > 0 {
				out.Messages = append(out.Messages, relaymodel.Message{Role: msg.Role, Content: parts})
			}
		}
	}
	// tools
	for _, t := range req.Tools {
		schema := t.InputSchema
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object"}`)
		}
		var params any
		_ = json.Unmarshal(schema, &params)
		out.Tools = append(out.Tools, relaymodel.Tool{
			Type: "function",
			Function: relaymodel.Function{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	if req.ToolChoice != nil {
		switch req.ToolChoice.Type {
		case "auto":
			out.ToolChoice = "auto"
		case "any":
			out.ToolChoice = "required"
		case "tool":
			out.ToolChoice = map[string]any{
				"type":     "function",
				"function": map[string]any{"name": req.ToolChoice.Name},
			}
		}
	}
	return out, nil
}

func hasToolResult(blocks []anthropicContentBlock) bool {
	for _, b := range blocks {
		if b.Type == "tool_result" {
			return true
		}
	}
	return false
}

// anthropicTextOf extracts plain text from `string | [{type:"text",text}]`.
func anthropicTextOf(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	if raw[0] == '"' {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			return s
		}
		return ""
	}
	var blocks []anthropicContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	texts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if b.Type == "text" {
			texts = append(texts, b.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func anthropicImageURL(src *anthropicImageSource) string {
	if src == nil {
		return ""
	}
	switch src.Type {
	case "base64":
		return fmt.Sprintf("data:%s;base64,%s", src.MediaType, src.Data)
	case "url":
		return src.Url
	}
	return ""
}

// ------------------------------------------------------------ handler ----

// anthropicBodyKey stashes the ORIGINAL Anthropic body across failover
// retries: the shared request-body cache (ctxkey.KeyRequestBody) is replaced
// with the converted OpenAI JSON, so the next retry attempt must find the
// pristine Anthropic body here to convert again.
const anthropicBodyKey = "one-api-next.anthropic_original_body"

func stashOriginalAnthropicBody(c *gin.Context) []byte {
	if v, ok := c.Get(anthropicBodyKey); ok {
		if raw, ok := v.([]byte); ok {
			return raw
		}
	}
	raw, err := common.GetRequestBody(c)
	if err != nil {
		return nil
	}
	c.Set(anthropicBodyKey, raw)
	return raw
}

// RelayMessagesHelper serves POST /v1/messages (Anthropic Messages API).
func RelayMessagesHelper(c *gin.Context) *relaymodel.ErrorWithStatusCode {
	requestBody := stashOriginalAnthropicBody(c)
	if len(requestBody) == 0 {
		return openai.ErrorWrapper(fmt.Errorf("empty request body"), "invalid_request_body", http.StatusBadRequest)
	}
	var areq anthropicRequest
	if err := json.Unmarshal(requestBody, &areq); err != nil {
		return openai.ErrorWrapper(err, "invalid_anthropic_request", http.StatusBadRequest)
	}
	if areq.Model == "" {
		return openai.ErrorWrapper(fmt.Errorf("model is required"), "invalid_request_error", http.StatusBadRequest)
	}
	if areq.MaxTokens <= 0 {
		return openai.ErrorWrapper(fmt.Errorf("max_tokens must be positive"), "invalid_request_error", 400)
	}
	if len(areq.Messages) == 0 {
		return openai.ErrorWrapper(fmt.Errorf("messages: at least one message is required"), "invalid_request_error", http.StatusBadRequest)
	}

	for _, message := range areq.Messages {
		if message.Role != "user" && message.Role != "assistant" {
			return openai.ErrorWrapper(fmt.Errorf("message role must be user or assistant"), "invalid_request_error", 400)
		}
		var content any
		if err := json.Unmarshal(message.Content, &content); err != nil {
			return openai.ErrorWrapper(err, "invalid_request_error", 400)
		}
		switch content.(type) {
		case string, []any:
		default:
			return openai.ErrorWrapper(fmt.Errorf("message content must be text or content blocks"), "invalid_request_error", 400)
		}
		copy := message
		copy.parseContent()
		for _, block := range copy.Blocks {
			switch block.Type {
			case "text", "image", "tool_use", "tool_result":
			default:
				return openai.ErrorWrapper(fmt.Errorf("unsupported content block: %s", block.Type), "invalid_request_error", 400)
			}
		}
	}
	for _, tool := range areq.Tools {
		if tool.Type != "" && tool.Type != "custom" {
			return openai.ErrorWrapper(fmt.Errorf("unsupported tool type: %s", tool.Type), "invalid_request_error", 400)
		}
	}
	var rawFields map[string]json.RawMessage
	_ = json.Unmarshal(requestBody, &rawFields)
	if raw := rawFields["thinking"]; len(raw) > 0 {
		var thinking struct {
			Type string `json:"type"`
		}
		if json.Unmarshal(raw, &thinking) != nil || thinking.Type != "disabled" {
			return openai.ErrorWrapper(fmt.Errorf("extended thinking is not supported"), "invalid_request_error", 400)
		}
	}
	if selectedModel := c.GetString(ctxkey.OriginalModel); selectedModel != "" {
		areq.Model = selectedModel
	}
	chatReq, err := anthropicToOpenAI(&areq)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_anthropic_request", http.StatusBadRequest)
	}
	converted, err := json.Marshal(chatReq)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// estimate input tokens for message_start.usage (the upstream only
	// reports usage at the end of the stream)
	inputEstimate := len(converted) / 4
	if inputEstimate < 1 {
		inputEstimate = 1
	}

	// swap body + path so the whole text pipeline (adaptors, URL builders,
	// billing, failover) runs in native chat mode. The shared body cache is
	// replaced too: every downstream parser (UnmarshalBodyReusable,
	// preserveExtraRequestFields, retry re-reads) must see the converted chat
	// JSON, while the original Anthropic body stays stashed above for the next
	// retry of THIS helper.
	c.Request.Body = io.NopCloser(bytes.NewBuffer(converted))
	c.Request.ContentLength = int64(len(converted))
	c.Set(ctxkey.KeyRequestBody, converted)
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = "/v1/chat/completions"
	defer func() { c.Request.URL.Path = originalPath }()

	w := newAnthropicResponseWriter(c.Writer, areq.Model, areq.Stream, areq.StopSequences, inputEstimate)
	c.Writer = w
	defer func() { c.Writer = w.underlying }()
	bizErr := RelayTextHelper(c)
	w.Finalize(bizErr != nil)
	c.Writer = w.underlying
	return bizErr
}

// RelayMessagesCountTokens serves POST /v1/messages/count_tokens with the
// gateway's conservative estimator (chars/4 — Anthropic tokenization is not
// available in-process; the count only guides client-side context trimming).
func RelayMessagesCountTokens(c *gin.Context) {
	requestBody, err := common.GetRequestBody(c)
	if err != nil || len(requestBody) == 0 {
		c.JSON(http.StatusBadRequest, anthropicErrorJSON("invalid_request_error", "empty request body"))
		return
	}
	var areq anthropicRequest
	if err := json.Unmarshal(requestBody, &areq); err != nil {
		c.JSON(http.StatusBadRequest, anthropicErrorJSON("invalid_request_error", err.Error()))
		return
	}
	if areq.Model == "" || len(areq.Messages) == 0 {
		c.JSON(400, anthropicErrorJSON("invalid_request_error", "model and messages are required"))
		return
	}
	tokens := len(requestBody) / 4
	if tokens < 1 {
		tokens = 1
	}
	c.JSON(http.StatusOK, gin.H{"input_tokens": tokens})
}

// --------------------------------------------------------- conversion ----

// anthropicErrorJSON builds the Anthropic error envelope.
func anthropicErrorJSON(errType, message string) gin.H {
	return gin.H{
		"type": "error",
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	}
}

// openAIErrorBodyToAnthropic converts an already-serialized OpenAI error body
// {"error":{...}} into the Anthropic envelope. Non-error bodies pass through.
func openAIErrorBodyToAnthropic(body []byte) []byte {
	var parsed struct {
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Error == nil {
		return nil
	}
	errType := "api_error"
	switch parsed.Error.Type {
	case "invalid_request_error", "authentication_error", "permission_error", "not_found_error", "rate_limit_error", "overloaded_error":
		errType = parsed.Error.Type
	}
	out, err := json.Marshal(map[string]any{
		"type": "error",
		"error": map[string]any{
			"type":    errType,
			"message": parsed.Error.Message,
		},
	})
	if err != nil {
		return nil
	}
	return out
}

// finishToStopReason maps OpenAI finish_reason to Anthropic stop_reason.
func finishToStopReason(reason string) string {
	switch reason {
	case "length":
		return "max_tokens"
	case "tool_calls", "function_call":
		return "tool_use"
	case "content_filter":
		return "refusal"
	default: // "stop" and everything unknown
		return "end_turn"
	}
}

// stopSequenceHit returns the matched stop sequence if text ends with one.
func stopSequenceHit(text string, seqs []string) (string, bool) {
	for _, s := range seqs {
		if s != "" && strings.HasSuffix(text, s) {
			return s, true
		}
	}
	return "", false
}

// convertOpenAIChatToAnthropic converts a full (non-stream) OpenAI chat
// completion JSON body into an Anthropic Messages response body.
func convertOpenAIChatToAnthropic(body []byte, originModel string, stopSequences []string, inputTokens int) []byte {
	var resp struct {
		Id      string `json:"id"`
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Role      string            `json:"role"`
				Content   any               `json:"content"`
				ToolCalls []relaymodel.Tool `json:"tool_calls"`
			} `json:"message"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage *relaymodel.Usage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil
	}
	var contentBlocks []map[string]any
	var fullText strings.Builder
	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		switch content := msg.Content.(type) {
		case string:
			if content != "" {
				fullText.WriteString(content)
				contentBlocks = append(contentBlocks, map[string]any{"type": "text", "text": content})
			}
		case []any:
			for _, part := range content {
				m, ok := part.(map[string]any)
				if !ok {
					continue
				}
				if t, _ := m["type"].(string); t == "text" {
					text, _ := m["text"].(string)
					fullText.WriteString(text)
					contentBlocks = append(contentBlocks, map[string]any{"type": "text", "text": text})
				}
			}
		}
		for _, tc := range msg.ToolCalls {
			input := parseToolArguments(tc.Function.Arguments)
			contentBlocks = append(contentBlocks, map[string]any{
				"type":  "tool_use",
				"id":    tc.Id,
				"name":  tc.Function.Name,
				"input": input,
			})
		}
	}
	if contentBlocks == nil {
		contentBlocks = []map[string]any{}
	}
	stopReason := "end_turn"
	var stopSequence any
	if len(resp.Choices) > 0 && resp.Choices[0].FinishReason != nil {
		stopReason = finishToStopReason(*resp.Choices[0].FinishReason)
	}
	if stopReason == "end_turn" {
		if seq, ok := stopSequenceHit(fullText.String(), stopSequences); ok {
			stopReason = "stop_sequence"
			stopSequence = seq
		}
	}
	usage := resp.Usage
	in := inputTokens
	out := 0
	if usage != nil {
		if usage.PromptTokens > 0 {
			in = usage.PromptTokens
		}
		out = usage.CompletionTokens
	}
	model := originModel
	if model == "" {
		model = resp.Model
	}
	result := map[string]any{
		"id":            "msg-" + strings.TrimPrefix(resp.Id, "chatcmpl-"),
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       contentBlocks,
		"stop_reason":   stopReason,
		"stop_sequence": stopSequence,
		"usage": map[string]any{
			"input_tokens":  in,
			"output_tokens": out,
		},
	}
	outJSON, err := json.Marshal(result)
	if err != nil {
		return nil
	}
	return outJSON
}

func parseToolArguments(args any) any {
	switch v := args.(type) {
	case nil:
		return map[string]any{}
	case string:
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			return parsed
		}
		return map[string]any{}
	default:
		return v
	}
}
