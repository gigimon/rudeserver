# HTTP Test Service (Go) — Design (2026-01-27)

## Goal
A single Go binary that helps test HTTP clients by returning controlled responses:
- Any HTTP status code
- Per-(URL+IP+method+protocol) rate limiting
- Optional delay
- Multiple protocol adapters via URL prefixes: `/http`, `/rest`, `/jsonrpc`
- "Smart URLs" with behavior controlled by query parameters

The repository was initialized with Git on January 27, 2026.

## Non-goals (MVP)
- Persistent storage
- Complex rule engines or config files
- Streaming, connection resets, redirects

## API Shape (MVP)
Path determines protocol and status code. Query determines behavior.

### Paths
- `/http/status/{code}`
- `/rest/status/{code}`
- `/jsonrpc/status/{code}`

If `{code}` is missing/invalid, default status is `200`.

### Query Parameters
- `rl`: rate limit in requests per second (RPS). Example: `rl=10`
- `burst`: burst size. Default: `max(1, rl)` when `rl` is set
- `delay`: response delay. Go duration format. Example: `delay=200ms`, `delay=1s`
- `body`: response body (string)
- `h`: response headers. Repeatable. Format: `Name:Value`
  - Example: `?h=Content-Type:application/json&h=X-Test:1`

### Examples
- `/http/status/429?rl=5&delay=250ms&body=too-many`
- `/rest/status/200?h=Content-Type:application/json&body={"ok":true}`
- `/jsonrpc/status/200`

## Protocol Behavior

### `/http`
- Accept any HTTP method
- Return status, headers, and body as specified
- No automatic header/body transformations

### `/rest`
- Accept any HTTP method
- Same behavior as `/http` for MVP
- (Optional future nicety: default `Content-Type: application/json` when body looks like JSON)

### `/jsonrpc`
- Accept only `POST`
  - Otherwise: `405 Method Not Allowed` and `Allow: POST`
- Validate request body as JSON-RPC:
  - Must parse as JSON object
  - Must include `jsonrpc: "2.0"`
  - Must include `id`
  - Otherwise: `400 Bad Request`
- Default status is `200` unless overridden by `/status/{code}`
- Response body (JSON-RPC):
  - `jsonrpc: "2.0"`
  - `id`: echo from request
  - `result`:
    - If `body` query param is set:
      - Try to parse it as JSON; if valid, use as JSON value
      - Otherwise, use it as a string
    - Else if request has `params`: use `params`
    - Else: echo entire request object
- Ensure `Content-Type: application/json` if not already set

## OpenAPI (Spec-First)
OpenAPI should be an explicit artifact that documents all three protocol prefixes, including JSON-RPC.

### Source of truth
- Keep the spec in `openapi/openapi.yaml`
- Optionally produce `openapi/openapi.json` from the YAML at build/start time

### Served endpoints
- `/openapi.yaml`: returns the YAML spec as static content
- `/openapi.json`: returns a JSON version of the spec

These endpoints are documentation-only and do not affect the scenario pipeline.

### What the spec should capture (MVP)
- Paths:
  - `/http/status/{code}`
  - `/rest/status/{code}`
  - `/jsonrpc/status/{code}`
- Common query parameters as reusable components:
  - `rl`, `burst`, `delay`, `body`, `h`
- Method constraints:
  - `/http` and `/rest`: accept any method (document the common set explicitly)
  - `/jsonrpc`: `POST` only
- Responses:
  - Use a `default` response plus documented examples for `200`, `400`, `405`, `429`

### Why spec-first here
- The API shape is unusual ("smart" behavior via query params), and explicit YAML is clearer
- It avoids noisy annotation-driven tooling in the handler code
- It allows clients and tests to generate SDKs or validate behavior even before full implementation

## Middleware Pipeline
Single shared pipeline across protocols:
1. Route + parse into `Scenario`
2. Resolve client IP
3. Rate limit (if configured)
4. Delay (if configured)
5. Protocol adapter builds response
6. Write response

### Why this order
- Rate limiting and delay are cross-cutting concerns and should run before response formatting
- Protocol adapters stay thin and focused on response construction

## Core Data Model
A small central model drives behavior.

```go
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
    NormalizedPath string // e.g., /status/429
    StatusCode     int    // default 200
    Delay          time.Duration
    RateLimit      *RateLimit
    Headers        http.Header
    Body           string
    ClientIP       string
}
```

## Rate Limiting Design
Use `golang.org/x/time/rate`.

### Key
Rate limit is isolated per protocol + method + path + client IP:

```text
key = protocol + "|" + method + "|" + normalizedPath + "|" + clientIP
```

This ensures different clients and different methods do not interfere.

### Client IP resolution
- If `X-Forwarded-For` is present, use its first IP
- Otherwise, parse from `RemoteAddr`

### Behavior on exceed
- Status: `429 Too Many Requests`
- Body: `rate limited` (unless a future rule changes this)

## Parsing Rules

### Path parsing
- Extract protocol from first segment
- Expect `status/{code}` next
- If `code` is invalid, keep default `200`

### Query parsing
- `rl`: parse as float > 0
- `burst`: parse as int > 0
- `delay`: parse with `time.ParseDuration`
- `h`: split on first `:` into header name/value; trim spaces
- Any invalid query value → `400 Bad Request`

## Error Contract
- Invalid query values: `400 Bad Request` with a short message
- JSON-RPC invalid JSON or invalid JSON-RPC shape: `400 Bad Request`
- JSON-RPC non-POST: `405 Method Not Allowed`
- Rate limit exceeded: `429 Too Many Requests`

## Suggested Package Structure

- `cmd/rudeserver/main.go`: entry point and HTTP server wiring
- `internal/httpserver/router.go`: protocol routing and pipeline assembly
- `internal/scenario/parser.go`: parse path/query into `Scenario`
- `internal/ip/ip.go`: client IP extraction
- `internal/ratelimit/ratelimit.go`: limiter store + middleware
- `internal/delay/delay.go`: delay middleware
- `internal/protocol/http.go`: HTTP adapter
- `internal/protocol/rest.go`: REST adapter (thin wrapper)
- `internal/protocol/jsonrpc.go`: JSON-RPC validation + response building
- `internal/openapi/handler.go`: serve `/openapi.yaml` and `/openapi.json`
- `openapi/openapi.yaml`: spec-first OpenAPI source of truth

## Testing Strategy (MVP)
Prefer table-driven tests with `net/http/httptest`.

### Unit tests
- Path parsing defaults and valid/invalid codes
- Query parsing (`rl`, `burst`, `delay`, repeated `h`, `body`)
- Rate-limit key composition
- Client IP extraction from `X-Forwarded-For` and `RemoteAddr`
- JSON-RPC validation and result selection logic

### Integration-ish handler tests
- `/http` accepts any method and returns status/body/headers
- Rate limiting yields `429`
- Delay is applied (assert min elapsed time with slack)
- `/openapi.yaml` and `/openapi.json` return specs with expected content type
- `/jsonrpc`:
  - POST + valid JSON-RPC returns 200 and echoes `id`
  - Non-POST returns 405
  - Invalid JSON or missing jsonrpc/id returns 400
  - `result` precedence: `body` → `params` → whole request

## Minimal CLI/Runtime Defaults
- Default listen address: `:8080`
- Timeouts (sane defaults): read/write/idle timeouts on the server
- No config file; everything controlled per-request

## Notes / Future Extensions
- Add `Retry-After` for 429
- Support more smart paths (e.g., `/status/{code}/delay/{d}`)
- Add chaos behaviors (disconnects, partial writes, chunked responses)
- Optional config layer for repeatable scenarios
- OpenAPI UI (e.g., Swagger UI) mounted at `/docs`
