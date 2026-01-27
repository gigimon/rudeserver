package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rudeserver/internal/ratelimit"
)

func TestRouterHTTP(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/http/status/418?body=hi&h=X-Test:1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 418 {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.String() != "hi" {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if rec.Header().Get("X-Test") != "1" {
		t.Fatalf("header = %q", rec.Header().Get("X-Test"))
	}
}

func TestRouterREST(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodPost, "/rest/status/201?body=ok", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRouterJSONRPCValid(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	body := `{"jsonrpc":"2.0","id":1,"params":{"a":1}}`
	req := httptest.NewRequest(http.MethodPost, "/jsonrpc/status/200", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}

	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json parse: %v", err)
	}
	if got["jsonrpc"] != "2.0" {
		t.Fatalf("jsonrpc = %v", got["jsonrpc"])
	}
	if got["id"] != float64(1) {
		t.Fatalf("id = %v", got["id"])
	}
	result, ok := got["result"].(map[string]any)
	if !ok {
		t.Fatalf("result type = %T", got["result"])
	}
	if result["a"] != float64(1) {
		t.Fatalf("result.a = %v", result["a"])
	}
}

func TestRouterJSONRPCRejectsNonPost(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/jsonrpc/status/200", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterJSONRPCRejectsInvalidJSON(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodPost, "/jsonrpc/status/200", strings.NewReader("nope"))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterJSONRPCRejectsMissingFields(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	body := `{"id":1}`
	req := httptest.NewRequest(http.MethodPost, "/jsonrpc/status/200", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRouterRateLimit(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/http/status/200?rl=1&burst=1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first status = %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d", rec2.Code)
	}
}

func TestRouterDelay(t *testing.T) {
	router := NewRouter(ratelimit.NewStore())
	req := httptest.NewRequest(http.MethodGet, "/http/status/200?delay=50ms", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	router.ServeHTTP(rec, req)
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Fatalf("elapsed = %v", elapsed)
	}
}
