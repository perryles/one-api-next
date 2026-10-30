package router

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProtocolRoutesRequireAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetRelayRouter(r)
	for _, path := range []string{"/v1/responses", "/v1/messages", "/v1/messages/count_tokens"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{"model":"test"}`)))
		if w.Code != 401 {
			t.Fatalf("%s returned %d: %s", path, w.Code, w.Body.String())
		}
		if strings.Contains(path, "messages") && !strings.Contains(w.Body.String(), `"type":"authentication_error"`) {
			t.Fatalf("wrong Messages envelope: %s", w.Body.String())
		}
	}
}
