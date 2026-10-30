package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/perryles/one-api-next/common/client"
	"github.com/perryles/one-api-next/common/config"
	"github.com/perryles/one-api-next/common/ctxkey"
	dbmodel "github.com/perryles/one-api-next/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResponsesMappedRelayAndBilling(t *testing.T) {
	setupPreConsumeTestDB(t)
	if err := dbmodel.DB.AutoMigrate(&dbmodel.ModelPrice{}, &dbmodel.GroupPrice{}, &dbmodel.Channel{}); err != nil {
		t.Fatal(err)
	}
	oldLog, oldApprox := config.LogConsumeEnabled, config.ApproximateTokenEnabled
	config.LogConsumeEnabled = false
	config.ApproximateTokenEnabled = true
	t.Cleanup(func() { config.LogConsumeEnabled = oldLog; config.ApproximateTokenEnabled = oldApprox })
	oldClient := client.HTTPClient
	client.HTTPClient = http.DefaultClient
	t.Cleanup(func() { client.HTTPClient = oldClient })
	uid := seedUser(t, 1_000_000)
	tid := seedToken(t, uid, 1_000_000)
	price := &dbmodel.ModelPrice{ModelName: "upstream-model", InputPrice: 10, OutputPrice: 20, CachedPrice: 2, BillingType: "token", Enabled: true}
	if err := dbmodel.DB.Create(price).Error; err != nil {
		t.Fatal(err)
	}
	dbmodel.InitModelPriceCache()
	dbmodel.InitGroupPriceCache()
	received := false
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer upstream-secret" {
			t.Errorf("bad upstream request: %s %s", r.URL.Path, r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload["model"] != "upstream-model" || payload["instructions"] != "Keep me" || payload["metadata"].(map[string]any)["custom"] != "yes" {
			t.Errorf("fields lost: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_mock","output":[],"usage":{"input_tokens":100,"input_tokens_details":{"cached_tokens":40},"output_tokens":20,"total_tokens":120}}`))
	}))
	defer upstream.Close()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"public-alias","input":"hi","instructions":"Keep me","metadata":{"custom":"yes"}}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Authorization", "Bearer upstream-secret")
	c.Set(ctxkey.Id, uid)
	c.Set(ctxkey.TokenId, tid)
	c.Set(ctxkey.ChannelID, "openai")
	c.Set(ctxkey.BaseURL, upstream.URL)
	c.Set(ctxkey.Group, "default")
	c.Set(ctxkey.ModelMapping, map[string]string{"public-alias": "upstream-model"})
	if err := RelayResponsesHelper(c); err != nil {
		t.Fatalf("relay failed: %+v", err)
	}
	if !received || w.Code != 200 || !strings.Contains(w.Body.String(), "resp_mock") {
		t.Fatalf("bad response: %d %s", w.Code, w.Body.String())
	}
	// 60 uncached input * 10 + 40 cached * 2 + 20 output * 20 = 1080 micro-yuan.
	if got := readUserQuota(t, uid); got != 1_000_000-1080 {
		t.Fatalf("quota=%d want=%d", got, 1_000_000-1080)
	}
	if got := readToken(t, tid).RemainQuota; got != 1_000_000-1080 {
		t.Fatalf("token quota=%d", got)
	}
}

func TestResponsesUpstreamFailureRefundsReservation(t *testing.T) {
	setupPreConsumeTestDB(t)
	if err := dbmodel.DB.AutoMigrate(&dbmodel.ModelPrice{}, &dbmodel.GroupPrice{}); err != nil {
		t.Fatal(err)
	}
	oldApprox := config.ApproximateTokenEnabled
	config.ApproximateTokenEnabled = true
	t.Cleanup(func() { config.ApproximateTokenEnabled = oldApprox })
	oldClient := client.HTTPClient
	client.HTTPClient = http.DefaultClient
	t.Cleanup(func() { client.HTTPClient = oldClient })
	uid := seedUser(t, 200_000)
	tid := seedToken(t, uid, 200_000)
	if err := dbmodel.DB.Create(&dbmodel.ModelPrice{ModelName: "failure-model", BillingType: dbmodel.BillingTypePerRequest, PerRequestPrice: 0.04, Enabled: true}).Error; err != nil {
		t.Fatal(err)
	}
	dbmodel.InitModelPriceCache()
	dbmodel.InitGroupPriceCache()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`{"error":{"message":"try again","type":"server_error"}}`))
	}))
	defer server.Close()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/responses", strings.NewReader(`{"model":"failure-model","input":"hi"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(ctxkey.Id, uid)
	c.Set(ctxkey.TokenId, tid)
	c.Set(ctxkey.ChannelID, "openai")
	c.Set(ctxkey.BaseURL, server.URL)
	c.Set(ctxkey.Group, "default")
	err := RelayResponsesHelper(c)
	if err == nil || err.StatusCode != 503 || c.Writer.Written() {
		t.Fatalf("upstream failure not retryable: %+v", err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if readUserQuota(t, uid) == 200_000 && readToken(t, tid).RemainQuota == 200_000 {
			return
		}
		time.Sleep(time.Millisecond * 5)
	}
	t.Fatalf("failed call retained reserved quota: user=%d token=%d", readUserQuota(t, uid), readToken(t, tid).RemainQuota)
}
