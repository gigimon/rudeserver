package openapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterServesOpenAPIArtifacts(t *testing.T) {
	mux := http.NewServeMux()
	if err := Register(mux); err != nil {
		t.Fatalf("register openapi: %v", err)
	}

	t.Run("yaml", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); ct == "" || !strings.Contains(ct, "yaml") {
			t.Fatalf("content-type = %q, want yaml", ct)
		}
		if rec.Body.Len() == 0 {
			t.Fatal("empty yaml body")
		}
	})

	t.Run("json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if ct := rec.Header().Get("Content-Type"); ct == "" || !strings.Contains(ct, "application/json") {
			t.Fatalf("content-type = %q, want application/json", ct)
		}
		if rec.Body.Len() == 0 {
			t.Fatal("empty json body")
		}
	})
}
