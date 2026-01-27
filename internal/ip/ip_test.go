package ip

import (
	"net/http"
	"testing"
)

func TestClientIPUsesXForwardedFor(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2")
	req.RemoteAddr = "192.0.2.10:1234"

	got := ClientIP(req)
	if got != "203.0.113.1" {
		t.Fatalf("ip = %q", got)
	}
}

func TestClientIPFallsBackToRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.10:1234"

	got := ClientIP(req)
	if got != "192.0.2.10" {
		t.Fatalf("ip = %q", got)
	}
}
