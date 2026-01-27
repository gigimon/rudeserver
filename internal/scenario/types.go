package scenario

import (
	"net/http"
	"time"
)

type Protocol string

const (
	ProtocolHTTP    Protocol = "http"
	ProtocolREST    Protocol = "rest"
	ProtocolJSONRPC Protocol = "jsonrpc"
)

type RateLimit struct {
	RPS   float64
	Burst int
}

type Scenario struct {
	Protocol       Protocol
	Method         string
	NormalizedPath string
	StatusCode     int
	Delay          time.Duration
	RateLimit      *RateLimit
	Headers        http.Header
	Body           string
}
