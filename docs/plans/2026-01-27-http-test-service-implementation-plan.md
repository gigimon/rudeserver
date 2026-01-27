# RudeServer MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the MVP HTTP test service in Go with smart URLs, per-(protocol|method|path|IP) rate limiting, JSON-RPC validation, and spec-first OpenAPI endpoints.

**Architecture:** A single `net/http` server with a shared middleware pipeline. The router parses protocol + status from the path, query params drive behavior, and thin protocol adapters finalize the response. OpenAPI is spec-first and served as static endpoints.

**Tech Stack:** Go (`net/http`, `httptest`), `golang.org/x/time/rate`

### Task 1: Initialize Go Module (Minimal Skeleton)

**Files:**
- Create: `go.mod`
- Create: `cmd/rudeserver/main.go`

**Step 1: Write the failing test**

Create: `cmd/rudeserver/main_test.go`

```go
package main

import "testing"

func TestMainBuilds(t *testing.T) {}
```

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL with “directory prefix . does not contain main module”

**Step 3: Write minimal implementation**

Create `go.mod` with module name `rudeserver` and current Go version.

Create `cmd/rudeserver/main.go` with a small `main()` that starts an HTTP server (it can serve a placeholder handler initially).

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for the new minimal module.

**Step 5: Commit**

```bash
git add go.mod cmd/rudeserver/main.go cmd/rudeserver/main_test.go
git commit -m "chore: initialize go module and main entrypoint"
```

### Task 2: Add Spec-First OpenAPI Artifacts and Endpoints

**Files:**
- Create: `openapi/openapi.yaml`
- Create: `internal/openapi/handler.go`
- Modify: `cmd/rudeserver/main.go`
- Test: `internal/openapi/handler_test.go`

**Step 1: Write the failing test**

Create: `internal/openapi/handler_test.go`

Test:
- `/openapi.yaml` returns 200 and `application/yaml` or `text/yaml`
- `/openapi.json` returns 200 and `application/json`

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL because the handler does not exist yet.

**Step 3: Write minimal implementation**

Create a spec-first OpenAPI file at `openapi/openapi.yaml` describing:
- `/http/status/{code}`
- `/rest/status/{code}`
- `/jsonrpc/status/{code}`
- Query params: `rl`, `burst`, `delay`, `body`, repeated `h`
- JSON-RPC `POST` only

Implement `internal/openapi/handler.go` that:
- Loads `openapi/openapi.yaml` at startup
- Serves it at `/openapi.yaml`
- Serves a JSON representation at `/openapi.json` (YAML-to-JSON conversion)

Wire endpoints in `cmd/rudeserver/main.go`.

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for the openapi handler tests.

**Step 5: Commit**

```bash
git add openapi/openapi.yaml internal/openapi/handler.go internal/openapi/handler_test.go cmd/rudeserver/main.go
git commit -m "feat: add spec-first openapi endpoints"
```

### Task 3: Implement Scenario Parsing (Path + Query)

**Files:**
- Create: `internal/scenario/types.go`
- Create: `internal/scenario/parser.go`
- Test: `internal/scenario/parser_test.go`

**Step 1: Write the failing test**

Create: `internal/scenario/parser_test.go`

Cover:
- Protocol parsing from prefix
- Status parsing from `/status/{code}` with default 200
- Query parsing: `rl`, `burst`, `delay`, repeated `h`, `body`
- Invalid query values return errors

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL due to missing parser.

**Step 3: Write minimal implementation**

Implement:
- `Scenario`, `Protocol`, `RateLimit`
- `Parse(r *http.Request) (Scenario, error)`

Keep it small and explicit; avoid premature abstractions.

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for parser tests.

**Step 5: Commit**

```bash
git add internal/scenario/types.go internal/scenario/parser.go internal/scenario/parser_test.go
git commit -m "feat: parse smart urls into scenarios"
```

### Task 4: Client IP Resolution + Rate Limiting Middleware

**Files:**
- Create: `internal/ip/ip.go`
- Create: `internal/ratelimit/limiter.go`
- Create: `internal/ratelimit/middleware.go`
- Test: `internal/ratelimit/middleware_test.go`

**Step 1: Write the failing test**

Create: `internal/ratelimit/middleware_test.go`

Cover:
- Key format: `protocol|method|path|ip`
- `X-Forwarded-For` precedence
- Limit exceed returns 429
- No `rl` means no limiting

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL due to missing middleware.

**Step 3: Write minimal implementation**

Implement:
- `ClientIP(r *http.Request) string`
- Limiter store keyed by `protocol|method|path|ip`
- Middleware that consults `Scenario.RateLimit`

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for rate limit tests.

**Step 5: Commit**

```bash
git add internal/ip/ip.go internal/ratelimit/*.go internal/ratelimit/middleware_test.go
git commit -m "feat: add per-client rate limiting middleware"
```

### Task 5: Delay Middleware

**Files:**
- Create: `internal/delay/middleware.go`
- Test: `internal/delay/middleware_test.go`

**Step 1: Write the failing test**

Create: `internal/delay/middleware_test.go`

Test:
- When `delay` is set, handler takes at least that long (with slack)

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL due to missing middleware.

**Step 3: Write minimal implementation**

Implement delay middleware that sleeps before calling the next handler when `Scenario.Delay > 0`.

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for delay tests.

**Step 5: Commit**

```bash
git add internal/delay/middleware.go internal/delay/middleware_test.go
git commit -m "feat: add delay middleware"
```

### Task 6: Protocol Adapters + Pipeline Router

**Files:**
- Create: `internal/protocol/http.go`
- Create: `internal/protocol/rest.go`
- Create: `internal/protocol/jsonrpc.go`
- Create: `internal/httpserver/router.go`
- Modify: `cmd/rudeserver/main.go`
- Test: `internal/httpserver/router_test.go`

**Step 1: Write the failing test**

Create: `internal/httpserver/router_test.go`

Cover:
- `/http` accepts any method and returns configured status/body/headers
- `/rest` behaves like `/http` for MVP
- `/jsonrpc`:
  - POST + valid JSON-RPC returns 200 with echoed `id`
  - Non-POST returns 405
  - Invalid JSON or missing jsonrpc/id returns 400
  - `result` precedence: `body` → `params` → whole request
- Rate limiting and delay still apply

**Step 2: Run test to verify it fails**

Run: `go test ./...`
Expected: FAIL due to missing router/adapters.

**Step 3: Write minimal implementation**

Implement a small router that:
- Parses the `Scenario`
- Resolves IP
- Applies rate limiting and delay
- Dispatches to a protocol adapter

Keep structure flat and readable; prefer a few files over many tiny packages.

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS for router tests and the full suite.

**Step 5: Commit**

```bash
git add internal/protocol/*.go internal/httpserver/router.go internal/httpserver/router_test.go cmd/rudeserver/main.go
git commit -m "feat: implement smart url pipeline and protocol adapters"
```

### Task 7: Final Verification + Docs Sync

**Files:**
- Modify: `docs/plans/2026-01-27-http-test-service-design.md`

**Step 1: Re-read the design and verify requirements**

Checklist:
- Smart URLs + query params
- Any method for `/http` and `/rest`
- JSON-RPC POST-only + validation + result rules
- Rate limit key includes protocol + method + path + IP
- OpenAPI spec-first endpoints exist

**Step 2: Run full verification**

Run: `go test ./...`
Expected: PASS

**Step 3: Update design doc if needed**

Only adjust design doc to reflect reality if implementation required it.

**Step 4: Commit**

```bash
git add docs/plans/2026-01-27-http-test-service-design.md
git commit -m "docs: align design with implementation details"
```

