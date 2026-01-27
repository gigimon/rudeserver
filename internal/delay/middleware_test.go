package delay

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWrapAppliesDelay(t *testing.T) {
	delay := 50 * time.Millisecond
	h := Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), delay)

	start := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	elapsed := time.Since(start)
	if elapsed < delay {
		t.Fatalf("elapsed = %v, want >= %v", elapsed, delay)
	}
}
