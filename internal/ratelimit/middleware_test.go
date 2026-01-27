package ratelimit

import (
	"testing"

	"rudeserver/internal/scenario"
)

func TestKeyFormat(t *testing.T) {
	sc := scenario.Scenario{
		Protocol:       scenario.ProtocolHTTP,
		Method:         "GET",
		NormalizedPath: "/status/200",
	}

	got := Key(sc, "203.0.113.1")
	want := "http|GET|/status/200|203.0.113.1"
	if got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

func TestAllowHonorsRateLimit(t *testing.T) {
	sc := scenario.Scenario{
		Protocol:       scenario.ProtocolHTTP,
		Method:         "GET",
		NormalizedPath: "/status/200",
		RateLimit: &scenario.RateLimit{
			RPS:   1,
			Burst: 1,
		},
	}

	store := NewStore()
	if !Allow(store, sc, "203.0.113.1") {
		t.Fatal("first allow should pass")
	}
	if Allow(store, sc, "203.0.113.1") {
		t.Fatal("second allow should be rate limited")
	}
}

func TestAllowWithoutRateLimit(t *testing.T) {
	sc := scenario.Scenario{
		Protocol:       scenario.ProtocolHTTP,
		Method:         "GET",
		NormalizedPath: "/status/200",
	}

	store := NewStore()
	if !Allow(store, sc, "203.0.113.1") {
		t.Fatal("allow without rate limit should pass")
	}
}
