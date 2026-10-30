package controller

// Anthropic Messages protocol conversion tests (relay/controller/messages.go).

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	relaymodel "github.com/perryles/one-api-next/relay/schema"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestAnthropicToOpenAIBasic(t *testing.T) {
	temp := 0.7
	req := &anthropicRequest{
		Model: "claude-sonnet-4-5",
		Messages: []anthropicMessage{
			{Role: "user", Content: json.RawMessage(`"你好"`)},
		},
		System:        json.RawMessage(`"You are helpful."`),
		MaxTokens:     1024,
		Temperature:   &temp,
		StopSequences: []string{"END"},
	}
	out, err := anthropicToOpenAI(req)
	if err != nil {
		t.Fatal(err)
	}
	if out.Model != "claude-sonnet-4-5" || out.MaxTokens != 1024 {
		t.Fatalf("basic fields: %+v", out)
	}
	if len(out.Messages) != 2 {
		t.Fatalf("system + user expected, got %d: %s", len(out.Messages), mustJSON(t, out.Messages))
	}
	if out.Messages[0].Role != "system" || out.Messages[0].Content != "You are helpful." {
		t.Fatalf("system message: %+v", out.Messages[0])
	}
	if out.Messages[1].Role != "user" || out.Messages[1].Content != "你好" {
		t.Fatalf("user message: %+v", out.Messages[1])
	}
	if stop, ok := out.Stop.([]string); !ok || len(stop) != 1 || stop[0] != "END" {
		t.Fatalf("stop sequences: %+v", out.Stop)
	}
	if out.Stream {
		t.Fatal("non-stream request must not set stream")
	}
	if out.StreamOptions != nil {
		t.Fatalf("non-stream request must not force include_usage: %+v", out.StreamOptions)
	}
}

func TestAnthropicToOpenAIToolRound(t *testing.T) {
	req := &anthropicRequest{
		Model: "claude-sonnet-4-5",
		Messages: []anthropicMessage{
			{Role: "user", Content: json.RawMessage(`"查一下北京天气"`)},
			{Role: "assistant", Content: json.RawMessage(`[
				{"type":"text","text":"我查一下"},
				{"type":"tool_use","id":"toolu_01","name":"get_weather","input":{"city":"北京"}}
			]`)},
			{Role: "user", Content: json.RawMessage(`[
				{"type":"tool_result","tool_use_id":"toolu_01","content":"晴 25℃"},
				{"type":"text","text":"再总结一句"}
			]`)},
		},
		Tools: []anthropicTool{{
			Name:        "get_weather",
			Description: "query weather",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		}},
		ToolChoice: &anthropicToolChoice{Type: "auto"},
		MaxTokens:  512,
	}
	out, err := anthropicToOpenAI(req)
	if err != nil {
		t.Fatal(err)
	}
	// user → assistant(text+tool_calls) → tool + user(text)
	var roles []string
	for _, m := range out.Messages {
		roles = append(roles, m.Role)
	}
	want := []string{"user", "assistant", "tool", "user"}
	if strings.Join(roles, ",") != strings.Join(want, ",") {
		t.Fatalf("roles: %v (msgs %s)", roles, mustJSON(t, out.Messages))
	}
	assistant := out.Messages[1]
	if assistant.ToolCalls[0].Id != "toolu_01" || assistant.ToolCalls[0].Function.Name != "get_weather" {
		t.Fatalf("tool call: %+v", assistant.ToolCalls)
	}
	if args, _ := assistant.ToolCalls[0].Function.Arguments.(string); args != `{"city":"北京"}` {
		t.Fatalf("tool arguments: %v", assistant.ToolCalls[0].Function.Arguments)
	}
	toolMsg := out.Messages[2]
	if toolMsg.ToolCallId != "toolu_01" || toolMsg.Content != "晴 25℃" {
		t.Fatalf("tool result: %+v", toolMsg)
	}
	if out.Messages[3].Content != "再总结一句" {
		t.Fatalf("trailing user text: %+v", out.Messages[3])
	}
	// tools → openai function tools
	if len(out.Tools) != 1 || out.Tools[0].Function.Name != "get_weather" {
		t.Fatalf("tools: %+v", out.Tools)
	}
	if out.ToolChoice != "auto" {
		t.Fatalf("tool choice: %+v", out.ToolChoice)
	}
}

func TestAnthropicToOpenAIImageAndStream(t *testing.T) {
	req := &anthropicRequest{
		Model:  "claude-sonnet-4-5",
		Stream: true,
		Messages: []anthropicMessage{
			{Role: "user", Content: json.RawMessage(`[
				{"type":"image","source":{"type":"base64","media_type":"image/png","data":"aGk="}},
				{"type":"text","text":"这是什么"}
			]`)},
		},
		MaxTokens: 64,
	}
	out, err := anthropicToOpenAI(req)
	if err != nil {
		t.Fatal(err)
	}
	if out.StreamOptions == nil || !out.StreamOptions.IncludeUsage {
		t.Fatalf("stream must force include_usage for billing: %+v", out.StreamOptions)
	}
	user := out.Messages[0]
	parts, ok := user.Content.([]relaymodel.MessageContent)
	if !ok || len(parts) != 2 {
		t.Fatalf("multimodal parts: %+v", user.Content)
	}
	if parts[0].Type != "image_url" || parts[0].ImageURL.Url != "data:image/png;base64,aGk=" {
		t.Fatalf("image part: %+v", parts[0])
	}
	if parts[1].Text != "这是什么" {
		t.Fatalf("text part: %+v", parts[1])
	}
}

func TestConvertOpenAIChatToAnthropic(t *testing.T) {
	body := []byte(`{
		"id": "chatcmpl-abc123",
		"model": "claude-sonnet-4-5",
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "你好，END"
			},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 11, "completion_tokens": 7, "total_tokens": 18}
	}`)
	out := convertOpenAIChatToAnthropic(body, "claude-sonnet-4-5", []string{"END"}, 5)
	if out == nil {
		t.Fatal("conversion returned nil")
	}
	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp["id"] != "msg-abc123" || resp["type"] != "message" || resp["role"] != "assistant" {
		t.Fatalf("envelope: %s", out)
	}
	if resp["stop_reason"] != "stop_sequence" {
		t.Fatalf("stop sequence must win over end_turn: %s", out)
	}
	blocks := resp["content"].([]any)
	if len(blocks) != 1 || blocks[0].(map[string]any)["text"] != "你好，END" {
		t.Fatalf("content: %s", out)
	}
	usage := resp["usage"].(map[string]any)
	if usage["input_tokens"] != float64(11) || usage["output_tokens"] != float64(7) {
		t.Fatalf("usage: %s", out)
	}
}

func TestConvertOpenAIChatToAnthropicToolUse(t *testing.T) {
	body := []byte(`{
		"id": "chatcmpl-t1",
		"model": "gpt-4o",
		"choices": [{
			"message": {
				"role": "assistant",
				"content": null,
				"tool_calls": [{"id":"call_1","type":"function","function":{"name":"lookup","arguments":"{\"q\":\"x\"}"}}]
			},
			"finish_reason": "tool_calls"
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
	}`)
	out := convertOpenAIChatToAnthropic(body, "claude-sonnet-4-5", nil, 3)
	var resp map[string]any
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatal(err)
	}
	if resp["stop_reason"] != "tool_use" {
		t.Fatalf("stop reason: %s", out)
	}
	block := resp["content"].([]any)[0].(map[string]any)
	if block["type"] != "tool_use" || block["id"] != "call_1" || block["name"] != "lookup" {
		t.Fatalf("tool block: %s", out)
	}
	if input := block["input"].(map[string]any); input["q"] != "x" {
		t.Fatalf("tool input not parsed from arguments string: %s", out)
	}
}

func TestAnthropicStreamTranslation(t *testing.T) {
	var sink bytes.Buffer
	w := &fakeGinWriter{buf: &sink}
	aw := newAnthropicResponseWriter(w, "claude-sonnet-4-5", true, []string{"STOP"}, 9)

	feed := func(line string) {
		_, _ = aw.Write([]byte("data: " + line + "\n\n"))
	}
	feed(`{"id":"chatcmpl-s1","model":"gpt-4o","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[{"index":0,"delta":{"content":"你好"},"finish_reason":null}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[{"index":0,"delta":{"content":"世界"},"finish_reason":null}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"call_9","function":{"name":"f","arguments":"{\"a\""}}]},"finish_reason":null}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"function":{"arguments":":1}"}}]},"finish_reason":null}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`)
	feed(`{"id":"chatcmpl-s1","choices":[],"usage":{"prompt_tokens":12,"completion_tokens":6,"total_tokens":18}}`)
	feed(`[DONE]`)
	aw.Finalize(false)

	out := sink.String()
	for _, want := range []string{
		"event: message_start",
		`"input_tokens":9`,
		"event: content_block_start", // text block
		"event: content_block_delta",
		`"text_delta"`,
		"event: content_block_stop", // text closes before tool block
		`"type":"tool_use"`,
		`"input_json_delta"`,
		`"stop_reason":"tool_use"`,
		`"output_tokens":6`,
		"event: message_stop",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("stream output missing %q:\n%s", want, out)
		}
	}
	if strings.Count(out, "event: message_start") != 1 {
		t.Fatalf("message_start must be emitted exactly once:\n%s", out)
	}
}

func TestAnthropicStreamStopSequence(t *testing.T) {
	var sink bytes.Buffer
	w := &fakeGinWriter{buf: &sink}
	aw := newAnthropicResponseWriter(w, "claude-sonnet-4-5", true, []string{"STOP"}, 3)
	feed := func(line string) { _, _ = aw.Write([]byte("data: " + line + "\n\n")) }
	feed(`{"id":"c1","choices":[{"index":0,"delta":{"role":"assistant"},"finish_reason":null}]}`)
	feed(`{"id":"c1","choices":[{"index":0,"delta":{"content":"done STOP"},"finish_reason":null}]}`)
	feed(`{"id":"c1","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`)
	feed(`{"id":"c1","choices":[],"usage":{"prompt_tokens":4,"completion_tokens":2,"total_tokens":6}}`)
	aw.Finalize(false)
	out := sink.String()
	if !strings.Contains(out, `"stop_reason":"stop_sequence"`) {
		t.Fatalf("stop_sequence detection:\n%s", out)
	}
	if !strings.Contains(out, `"stop_sequence":"STOP"`) {
		t.Fatalf("stop_sequence value:\n%s", out)
	}
}

func TestAnthropicErrorEnvelope(t *testing.T) {
	converted := openAIErrorBodyToAnthropic([]byte(`{"error":{"message":"quota exceeded","type":"invalid_request_error"}}`))
	if converted == nil {
		t.Fatal("nil conversion")
	}
	var resp map[string]any
	if err := json.Unmarshal(converted, &resp); err != nil {
		t.Fatal(err)
	}
	if resp["type"] != "error" {
		t.Fatalf("envelope: %s", converted)
	}
	inner := resp["error"].(map[string]any)
	if inner["type"] != "invalid_request_error" || inner["message"] != "quota exceeded" {
		t.Fatalf("inner: %s", converted)
	}
	// non-error bodies pass through as nil
	if openAIErrorBodyToAnthropic([]byte(`{"id":"x","choices":[]}`)) != nil {
		t.Fatal("non-error body must not convert")
	}
}

func TestAnthropicNonStreamWriterConverts(t *testing.T) {
	var sink bytes.Buffer
	w := &fakeGinWriter{buf: &sink}
	aw := newAnthropicResponseWriter(w, "claude-sonnet-4-5", false, nil, 4)
	aw.WriteHeader(200)
	_, _ = aw.Write([]byte(`{"id":"chatcmpl-n1","model":"m","choices":[{"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1,"total_tokens":4}}`))
	aw.Finalize(false)
	out := sink.String()
	if !strings.Contains(out, `"type":"message"`) || !strings.Contains(out, `"input_tokens":3`) {
		t.Fatalf("non-stream conversion: %s", out)
	}
	// error body → anthropic envelope
	sink.Reset()
	aw = newAnthropicResponseWriter(w, "claude-sonnet-4-5", false, nil, 4)
	aw.WriteHeader(429)
	_, _ = aw.Write([]byte(`{"error":{"message":"rate limited","type":"rate_limit_error"}}`))
	aw.Finalize(true)
	if !strings.Contains(sink.String(), `"type":"error"`) {
		t.Fatalf("error conversion: %s", sink.String())
	}
}
