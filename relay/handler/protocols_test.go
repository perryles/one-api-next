package controller

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/common/config"
	"github.com/perryles/one-api-next/relay/meta"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponsesNonStreamUsage(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"id":"resp_1","usage":{"input_tokens":100,"input_tokens_details":{"cached_tokens":40},"output_tokens":20,"total_tokens":120},"output":[]}`
	resp := &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
	usage, err := responsesDoResponse(c, resp, &meta.Meta{})
	if err != nil || usage.PromptTokens != 100 || usage.CompletionTokens != 20 || usage.PromptTokensDetails.CachedTokens != 40 {
		t.Fatalf("usage=%+v err=%+v", usage, err)
	}
	if w.Body.String() != body {
		t.Fatalf("response modified: %s", w.Body.String())
	}
}

func TestResponsesInvalidJSONDoesNotWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	_, err := responsesDoResponse(c, &http.Response{Header: make(http.Header), Body: io.NopCloser(strings.NewReader("broken"))}, &meta.Meta{})
	if err == nil || c.Writer.Written() {
		t.Fatalf("invalid response must be rejected before headers: %+v", err)
	}
}

func TestResponsesStreamUsageAndTruncation(t *testing.T) {
	previous := config.ApproximateTokenEnabled
	config.ApproximateTokenEnabled = true
	t.Cleanup(func() { config.ApproximateTokenEnabled = previous })
	for _, terminal := range []bool{true, false} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
		body := "event: response.output_text.delta\ndata:{\"type\":\"response.output_text.delta\",\"delta\":\"hello\"}\n\n"
		if terminal {
			body += "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"usage\":{\"input_tokens\":3,\"output_tokens\":2,\"total_tokens\":5}}}\n\n"
		}
		usage, err := responsesStreamHandler(c, &http.Response{Body: io.NopCloser(strings.NewReader(body))}, &meta.Meta{IsStream: true, ActualModelName: "gpt-4", PromptTokens: 3})
		if err != nil || usage == nil {
			t.Fatalf("usage=%+v err=%+v", usage, err)
		}
		if terminal && (usage.TotalTokens != 5 || w.Body.String() != body) {
			t.Fatalf("bad terminal response: %s usage=%+v", w.Body.String(), usage)
		}
		if !terminal && !strings.Contains(w.Body.String(), "upstream_stream_interrupted") {
			t.Fatalf("truncated stream was silently accepted: %s", w.Body.String())
		}
	}
}

func TestAnthropicFailureBeforeHeadersRemainsRetryable(t *testing.T) {
	for _, stream := range []bool{true, false} {
		var sink bytes.Buffer
		underlying := &fakeGinWriter{buf: &sink}
		writer := newAnthropicResponseWriter(underlying, "claude", stream, nil, 1)
		writer.Finalize(true)
		if underlying.Written() || sink.Len() != 0 {
			t.Fatalf("failure wrote an empty successful response: %s", sink.String())
		}
	}
}

func TestMessagesCountTokensValidation(t *testing.T) {
	for _, body := range []string{`{}`, `broken`, `{"model":"claude","messages":[{"role":"user","content":"hello"}]}`} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/messages/count_tokens", strings.NewReader(body))
		RelayMessagesCountTokens(c)
		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(body, "hello") {
			if w.Code != 200 || result["input_tokens"].(float64) < 1 {
				t.Fatal(w.Body.String())
			}
		} else if w.Code != 400 || result["type"] != "error" {
			t.Fatal(w.Body.String())
		}
	}
}

func TestResponsesRejectsUnimplementedStatefulModes(t *testing.T) {
	for _, body := range []string{`{"model":"test","input":"hi","previous_response_id":"resp_previous"}`, `{"model":"test","input":"hi","background":true}`} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(body))
		err := RelayResponsesHelper(c)
		if err == nil || err.StatusCode != 400 || c.Writer.Written() {
			t.Fatalf("stateful mode accepted: %+v", err)
		}
	}
}

func TestMessagesTruncatedStreamDoesNotReportSuccess(t *testing.T) {
	var sink bytes.Buffer
	underlying := &fakeGinWriter{buf: &sink}
	writer := newAnthropicResponseWriter(underlying, "claude", true, nil, 1)
	_, _ = writer.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"partial\"}}]}\n\n"))
	writer.Finalize(false)
	if !strings.Contains(sink.String(), "event: error") || strings.Contains(sink.String(), "event: message_stop") {
		t.Fatalf("truncated response reported as success: %s", sink.String())
	}
}
