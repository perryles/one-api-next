package controller

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/common"
	"github.com/perryles/one-api-next/common/config"
	"github.com/perryles/one-api-next/common/logger"
	dbmodel "github.com/perryles/one-api-next/model"
	"github.com/perryles/one-api-next/relay"
	"github.com/perryles/one-api-next/relay/adaptor/openai"
	"github.com/perryles/one-api-next/relay/billing"
	billingratio "github.com/perryles/one-api-next/relay/billing/ratio"
	"github.com/perryles/one-api-next/relay/meta"
	"github.com/perryles/one-api-next/relay/registry"
	"github.com/perryles/one-api-next/relay/schema"
	"io"
	"net/http"
	"strings"
)

// responsesRequest is the minimal shape needed from an OpenAI Responses API
// request (POST /v1/responses). The full body is relayed to the upstream
// unchanged; these fields only drive routing, model mapping and billing.
type responsesRequest struct {
	Model           string          `json:"model"`
	Stream          bool            `json:"stream"`
	Instructions    string          `json:"instructions"`
	Input           json.RawMessage `json:"input"`
	MaxOutputTokens int             `json:"max_output_tokens"`
	// Stateful continuations require channel binding and are rejected for now.
	PreviousResponseId string `json:"previous_response_id"`
	Store              bool   `json:"store"`
}

// responsesUsage mirrors the usage object in a Responses API result.
type responsesUsage struct {
	InputTokens        int                        `json:"input_tokens"`
	InputTokensDetails *model.PromptTokensDetails `json:"input_tokens_details"`
	OutputTokens       int                        `json:"output_tokens"`
	TotalTokens        int                        `json:"total_tokens"`
}

// responsesResult is used for both the non-stream response object and the
// response object embedded in the `response.completed` stream event.
type responsesResult struct {
	Id     string          `json:"id"`
	Usage  *responsesUsage `json:"usage"`
	Output []struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

// responsesStreamEvent is one SSE event of a streamed Responses API call.
type responsesStreamEvent struct {
	Type     string           `json:"type"`
	Delta    string           `json:"delta"`
	Response *responsesResult `json:"response"`
}

func RelayResponsesHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	raw, err := common.GetRequestBody(c)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_request_body", 400)
	}
	var req responsesRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return openai.ErrorWrapper(err, "invalid_responses_request", 400)
	}
	if req.Model == "" {
		return openai.ErrorWrapper(errors.New("model is required"), "invalid_responses_request", 400)
	}
	if req.MaxOutputTokens < 0 || len(req.Input) == 0 || string(req.Input) == "null" {
		return openai.ErrorWrapper(errors.New("input is required and max_output_tokens cannot be negative"), "invalid_responses_request", 400)
	}
	if req.PreviousResponseId != "" {
		return openai.ErrorWrapper(errors.New("previous_response_id is not supported; send conversation history in input"), "unsupported_parameter", 400)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return openai.ErrorWrapper(err, "invalid_responses_request", 400)
	}
	if string(fields["background"]) == "true" {
		return openai.ErrorWrapper(errors.New("background responses are not supported"), "unsupported_parameter", 400)
	}
	if !registry.SupportsResponses(meta.ChannelID) {
		return openai.ErrorWrapper(errors.New("responses requires an OpenAI-compatible channel"), "responses_not_supported", 400)
	}
	meta.IsStream = req.Stream
	meta.OriginModelName = req.Model
	meta.ActualModelName, _ = getMappedModelName(req.Model, meta.ModelMapping)
	fields["model"], _ = json.Marshal(meta.ActualModelName)
	raw, err = json.Marshal(fields)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_responses_request", 400)
	}
	price, err := billingratio.GetModelPrice(meta.ActualModelName, meta.OriginModelName)
	if err != nil {
		return openai.ErrorWrapper(err, "model_price_not_found", 422)
	}
	discount := billingratio.GetGroupDiscount(meta.Group, meta.ActualModelName, meta.OriginModelName)
	meta.PromptTokens = openai.CountTokenInput(req.Instructions+string(req.Input), meta.ActualModelName)
	request := &model.GeneralOpenAIRequest{Model: meta.ActualModelName, Stream: req.Stream, MaxTokens: req.MaxOutputTokens}
	var reserved int64
	var bizErr *model.ErrorWithStatusCode
	if price.BillingType == dbmodel.BillingTypePerRequest {
		reserved, bizErr = preConsumeExactQuota(ctx, billingratio.CalculatePerRequestQuota(price.PerRequestPrice, 1, 1, discount), meta)
	} else {
		ratio := (price.InputPrice + price.OutputPrice) / 2 / billingratio.Million * config.QuotaPerUnit
		if ratio == 0 {
			ratio = 1
		}
		reserved, bizErr = preConsumeQuota(ctx, request, meta.PromptTokens, ratio, meta)
	}
	if bizErr != nil {
		return bizErr
	}
	settled := false
	defer func() {
		if !settled {
			billing.ReturnPreConsumedQuota(ctx, reserved, meta.TokenId)
		}
	}()
	adapter := relay.GetAdaptorByChannelID(meta.ChannelID)
	if adapter == nil {
		return openai.ErrorWrapper(errors.New("invalid channel"), "invalid_channel", 400)
	}
	adapter.Init(meta)
	resp, err := adapter.DoRequest(c, meta, bytes.NewReader(raw))
	if err != nil {
		return openai.ErrorWrapper(err, "do_request_failed", 502)
	}
	defer resp.Body.Close()
	if isErrorHappened(meta, resp) {
		return RelayErrorHandler(resp)
	}
	usage, responseErr := responsesDoResponse(c, resp, meta)
	if responseErr != nil {
		return responseErr
	}
	postConsumeQuota(ctx, usage, meta, request, reserved, price, discount, false)
	settled = true
	return nil
}

// responsesDoResponse forwards the upstream result to the client and extracts
// the billable usage.
func responsesDoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode) {
	if meta.IsStream {
		return responsesStreamHandler(c, resp, meta)
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, openai.ErrorWrapper(err, "read_response_failed", http.StatusInternalServerError)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	var result responsesResult
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, openai.ErrorWrapper(err, "invalid_upstream_response", 502)
	}

	c.Data(http.StatusOK, contentType, responseBody)
	if result.Usage != nil {
		return responsesUsageToModel(result.Usage), nil
	}
	// upstream returned no usage: fall back to a conservative estimate
	var output strings.Builder
	for _, item := range result.Output {
		for _, part := range item.Content {
			output.WriteString(part.Text)
		}
	}
	return openai.ResponseText2Usage(output.String(), meta.ActualModelName, meta.PromptTokens), nil
}

// responsesStreamHandler pipes the SSE stream through unchanged (events keep
// their original framing) while capturing the usage carried by the final
// `response.completed` event. If the upstream never reports usage, it falls
// back to estimating from the accumulated output text.
func responsesStreamHandler(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode) {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	// response.completed carries the whole response object in one SSE line
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)

	var usage *model.Usage
	var responseText strings.Builder
	// F-2: a well-formed Responses stream always ends with a terminal event
	// (response.completed / response.incomplete / response.failed) or the SSE
	// terminator [DONE]. A clean FIN without either is a truncated stream.
	terminalSeen := false

	common.SetEventStreamHeaders(c)
	for scanner.Scan() {
		line := scanner.Text()
		if _, err := fmt.Fprintln(c.Writer, line); err != nil {
			break
		}
		c.Writer.Flush()

		payload, ok := stripSSEDataPrefix(line)
		if !ok {
			continue
		}
		if payload == "[DONE]" {
			terminalSeen = true
			continue
		}
		var event responsesStreamEvent
		if err := json.Unmarshal([]byte(payload), &event); err != nil {
			continue // non-JSON or irrelevant event, already forwarded as-is
		}
		switch event.Type {
		case "response.completed", "response.incomplete", "response.failed":
			terminalSeen = true
			if event.Response != nil {

				if event.Response.Usage != nil {
					usage = responsesUsageToModel(event.Response.Usage)
				}
			}
		case "response.output_text.delta", "response.function_call_arguments.delta":
			responseText.WriteString(event.Delta)
		}
	}
	// F-2: an upstream that drops the connection mid-stream must not leave the
	// client with a silent truncated 200 — neither an I/O error nor a clean FIN
	// without a terminal event is acceptable. Emit a terminating error event and
	// the SSE terminator; the usage produced so far is still settled below.
	if scanErr := scanner.Err(); scanErr != nil || !terminalSeen {
		reason := "upstream stream interrupted"
		if scanErr != nil {
			reason = "upstream stream interrupted: " + scanErr.Error()
		} else {
			reason = "upstream stream ended without a terminal event (clean FIN)"
		}
		logger.Errorf(c.Request.Context(), "responses %s (channel #%d, model %s)", reason, meta.ChannelId, meta.ActualModelName)
		// json.Marshal keeps the SSE payload valid whatever the error text is
		errorPayload, _ := json.Marshal(map[string]interface{}{
			"type": "error",
			"error": map[string]string{
				"message": reason,
				"type":    "upstream_error",
				"code":    "upstream_stream_interrupted",
			},
		})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errorPayload)
		fmt.Fprint(c.Writer, "data: [DONE]\n\n")
		c.Writer.Flush()
	}
	if usage == nil {
		usage = openai.ResponseText2Usage(responseText.String(), meta.ActualModelName, meta.PromptTokens)
	}
	return usage, nil
}

func responsesUsageToModel(u *responsesUsage) *model.Usage {
	return &model.Usage{
		PromptTokens:        u.InputTokens,
		PromptTokensDetails: u.InputTokensDetails,
		CompletionTokens:    u.OutputTokens,
		TotalTokens:         u.TotalTokens,
	}
}

// stripSSEDataPrefix returns the JSON payload of an SSE `data:` line.
func stripSSEDataPrefix(line string) (string, bool) {
	if strings.HasPrefix(line, "data:") {
		return strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "), "\r"), true
	}
	return "", false
}

// plainResponsesInput returns the input as a plain string when it is one
// (Responses API allows `input` to be a string or an array of content items).
func plainResponsesInput(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	return "", false
}
