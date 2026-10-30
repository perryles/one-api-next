package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/model"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMessagesAPIKeyAuthentication(t *testing.T) {
	seedUser(t, "messagesclient", model.RoleCommonUser, "messagesclientkey", true)
	for _, path := range []string{"/v1/messages", "/v1/messages/count_tokens", "/v1/responses"} {
		r := gin.New()
		r.POST(path, TokenAuth(), func(c *gin.Context) { c.Status(204) })
		req := httptest.NewRequest("POST", path, strings.NewReader(`{"model":"claude"}`))
		req.Header.Set("x-api-key", "sk-messagesclientkey")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if path == "/v1/responses" {
			if w.Code != 401 {
				t.Fatalf("x-api-key leaked to other protocols: %d", w.Code)
			}
		} else if w.Code != 204 {
			t.Fatalf("Messages API key rejected: %d %s", w.Code, w.Body.String())
		}
	}
	r := gin.New()
	r.POST("/v1/messages", TokenAuth())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{"model":"claude"}`)))
	if w.Code != 401 || !strings.Contains(w.Body.String(), `"type":"authentication_error"`) || !strings.Contains(w.Body.String(), `"type":"error"`) {
		t.Fatalf("wrong auth error: %d %s", w.Code, w.Body.String())
	}
}
