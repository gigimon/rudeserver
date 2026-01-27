package scenario

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestParseRequestHTTP(t *testing.T) {
	u := &url.URL{Path: "/http/status/404"}
	q := u.Query()
	q.Set("rl", "10")
	q.Set("burst", "2")
	q.Set("delay", "200ms")
	q.Set("body", "hello")
	q.Add("h", "Content-Type:application/json")
	q.Add("h", "X-Test:1")
	u.RawQuery = q.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	got, err := ParseRequest(req)
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}

	if got.Protocol != ProtocolHTTP {
		t.Fatalf("protocol = %q", got.Protocol)
	}
	if got.StatusCode != 404 {
		t.Fatalf("status = %d", got.StatusCode)
	}
	if got.NormalizedPath != "/status/404" {
		t.Fatalf("normalized path = %q", got.NormalizedPath)
	}
	if got.Delay != 200*time.Millisecond {
		t.Fatalf("delay = %v", got.Delay)
	}
	if got.RateLimit == nil || got.RateLimit.RPS != 10 || got.RateLimit.Burst != 2 {
		t.Fatalf("rate limit = %+v", got.RateLimit)
	}
	if got.Body != "hello" {
		t.Fatalf("body = %q", got.Body)
	}
	if got.Headers.Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q", got.Headers.Get("Content-Type"))
	}
	if got.Headers.Get("X-Test") != "1" {
		t.Fatalf("x-test = %q", got.Headers.Get("X-Test"))
	}
}

func TestParseRequestDefaultsStatusOnInvalidCode(t *testing.T) {
	u := &url.URL{Path: "/rest/status/abc"}
	req := &http.Request{Method: http.MethodGet, URL: u}

	got, err := ParseRequest(req)
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if got.StatusCode != 200 {
		t.Fatalf("status = %d", got.StatusCode)
	}
}

func TestParseRequestDefaultStatusOnMissingCode(t *testing.T) {
	u := &url.URL{Path: "/http/anything"}
	req := &http.Request{Method: http.MethodGet, URL: u}

	got, err := ParseRequest(req)
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if got.StatusCode != 200 {
		t.Fatalf("status = %d", got.StatusCode)
	}
}

func TestParseRequestInvalidRateLimit(t *testing.T) {
	u := &url.URL{Path: "/http/status/200"}
	q := u.Query()
	q.Set("rl", "nope")
	u.RawQuery = q.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	_, err := ParseRequest(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRequestInvalidDelay(t *testing.T) {
	u := &url.URL{Path: "/http/status/200"}
	q := u.Query()
	q.Set("delay", "nope")
	u.RawQuery = q.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	_, err := ParseRequest(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRequestInvalidHeader(t *testing.T) {
	u := &url.URL{Path: "/http/status/200"}
	q := u.Query()
	q.Add("h", "NoColon")
	u.RawQuery = q.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	_, err := ParseRequest(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRequestUnknownProtocol(t *testing.T) {
	u := &url.URL{Path: "/unknown/status/200"}
	req := &http.Request{Method: http.MethodGet, URL: u}
	_, err := ParseRequest(req)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRequestDefaultBurst(t *testing.T) {
	u := &url.URL{Path: "/http/status/200"}
	q := u.Query()
	q.Set("rl", "2.5")
	u.RawQuery = q.Encode()

	req := &http.Request{Method: http.MethodGet, URL: u}
	got, err := ParseRequest(req)
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	if got.RateLimit == nil || got.RateLimit.Burst != 3 {
		t.Fatalf("burst = %v", got.RateLimit)
	}
}
