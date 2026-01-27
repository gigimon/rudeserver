package scenario

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ParseRequest(r *http.Request) (Scenario, error) {
	if r == nil || r.URL == nil {
		return Scenario{}, fmt.Errorf("request is nil")
	}

	protocol, normalizedPath, status := parsePath(r.URL.Path)
	if protocol == "" {
		return Scenario{}, fmt.Errorf("unsupported protocol")
	}

	q := r.URL.Query()
	delay, err := parseDelay(q.Get("delay"))
	if err != nil {
		return Scenario{}, err
	}

	headers, err := parseHeaders(q["h"])
	if err != nil {
		return Scenario{}, err
	}

	rateLimit, err := parseRateLimit(q.Get("rl"), q.Get("burst"))
	if err != nil {
		return Scenario{}, err
	}

	body := q.Get("body")

	return Scenario{
		Protocol:       protocol,
		Method:         r.Method,
		NormalizedPath: normalizedPath,
		StatusCode:     status,
		Delay:          delay,
		RateLimit:      rateLimit,
		Headers:        headers,
		Body:           body,
	}, nil
}

func parsePath(path string) (Protocol, string, int) {
	trimmed := strings.TrimPrefix(path, "/")
	segments := strings.Split(trimmed, "/")
	if len(segments) == 0 || segments[0] == "" {
		return "", "", 200
	}

	var protocol Protocol
	switch segments[0] {
	case string(ProtocolHTTP):
		protocol = ProtocolHTTP
	case string(ProtocolREST):
		protocol = ProtocolREST
	case string(ProtocolJSONRPC):
		protocol = ProtocolJSONRPC
	default:
		return "", "", 200
	}

	normalized := "/"
	if len(segments) > 1 {
		normalized = "/" + strings.Join(segments[1:], "/")
	}

	status := 200
	if len(segments) >= 3 && segments[1] == "status" {
		if code, err := strconv.Atoi(segments[2]); err == nil {
			status = code
		}
	}

	return protocol, normalized, status
}

func parseDelay(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid delay")
	}
	return parsed, nil
}

func parseHeaders(values []string) (http.Header, error) {
	headers := make(http.Header)
	for _, raw := range values {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header")
		}
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" {
			return nil, fmt.Errorf("invalid header name")
		}
		headers.Add(name, value)
	}
	return headers, nil
}

func parseRateLimit(rlRaw string, burstRaw string) (*RateLimit, error) {
	if rlRaw == "" && burstRaw == "" {
		return nil, nil
	}
	if rlRaw == "" && burstRaw != "" {
		return nil, fmt.Errorf("burst requires rl")
	}

	rps, err := strconv.ParseFloat(rlRaw, 64)
	if err != nil || rps <= 0 {
		return nil, fmt.Errorf("invalid rl")
	}

	burst := 0
	if burstRaw != "" {
		parsed, err := strconv.Atoi(burstRaw)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("invalid burst")
		}
		burst = parsed
	} else {
		burst = int(math.Ceil(rps))
		if burst < 1 {
			burst = 1
		}
	}

	return &RateLimit{RPS: rps, Burst: burst}, nil
}
