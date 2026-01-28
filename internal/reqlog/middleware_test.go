package reqlog

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareCapturesAndPassesThrough(t *testing.T) {
	store := NewStore(10)

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Fatalf("downstream body = %q", string(body))
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(201)
		_, _ = w.Write([]byte("world"))
	})

	wrapped := Middleware(store, h)

	req := httptest.NewRequest(http.MethodPost, "/http/status/201?x=1", bytes.NewBufferString("hello"))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	entries := store.List()
	if len(entries) != 1 {
		t.Fatalf("entries = %d", len(entries))
	}
	entry := entries[0]
	if entry.Status != 201 {
		t.Fatalf("status = %d", entry.Status)
	}
	if string(entry.ReqBody) != "hello" {
		t.Fatalf("req body = %q", string(entry.ReqBody))
	}
	if string(entry.ResBody) != "world" {
		t.Fatalf("res body = %q", string(entry.ResBody))
	}
	if entry.ContentType != "text/plain" {
		t.Fatalf("content-type = %q", entry.ContentType)
	}
}

func TestMiddlewareExcludesUIAndOpenAPI(t *testing.T) {
	store := NewStore(10)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	wrapped := Middleware(store, h)

	paths := []string{"/", "/ui/app.js", "/openapi.json", "/openapi.yaml"}
	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
	}
	if len(store.List()) != 0 {
		t.Fatalf("expected no entries")
	}
}

func TestMiddlewareTruncatesBodies(t *testing.T) {
	store := NewStore(10)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(bytes.Repeat([]byte("b"), maxBodyBytes+10))
	})
	wrapped := Middleware(store, h)

	req := httptest.NewRequest(http.MethodPost, "/http/status/200", bytes.NewBuffer(bytes.Repeat([]byte("a"), maxBodyBytes+10)))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	entry := store.List()[0]
	if !entry.ReqTruncated || !entry.ResTruncated {
		t.Fatalf("expected truncation flags")
	}
	if len(entry.ReqBody) != maxBodyBytes || len(entry.ResBody) != maxBodyBytes {
		t.Fatalf("unexpected body sizes")
	}
}

func TestMiddlewareSetsProtocol(t *testing.T) {
	store := NewStore(10)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	wrapped := Middleware(store, h)

	req := httptest.NewRequest(http.MethodGet, "/jsonrpc/status/200", strings.NewReader("{}"))
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	entry := store.List()[0]
	if entry.Protocol != "jsonrpc" {
		t.Fatalf("protocol = %q", entry.Protocol)
	}
}
