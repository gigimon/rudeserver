# RudeServer — Agent Guide

## Project Overview
RudeServer is a Go HTTP test service for exercising HTTP clients under controlled conditions.
It supports:
- Arbitrary HTTP status codes
- Per-(protocol + method + path + client IP) rate limiting
- Optional response delay
- Protocol adapters via URL prefixes: `/http`, `/rest`, `/jsonrpc`
- Spec-first OpenAPI documentation

The MVP is intentionally small and driven by “smart URLs” plus query parameters.

## Current State (January 27, 2026)
There is no implementation yet. The design document is the source of truth.
- Read: `docs/plans/2026-01-27-http-test-service-design.md`

## Where To Look First
When starting work, read these in order:
1. `docs/plans/2026-01-27-http-test-service-design.md`
2. This file: `AGENTS.md`

## Key Contracts (MVP)
API shape:
- Paths: `/http/status/{code}`, `/rest/status/{code}`, `/jsonrpc/status/{code}`
- Query params: `rl`, `burst`, `delay`, `body`, repeated `h=Name:Value`

Protocol rules:
- `/http` and `/rest` accept any HTTP method
- `/jsonrpc` accepts only `POST`
- Default status is `200` when code is missing or invalid

JSON-RPC behavior:
- Validate `jsonrpc: "2.0"` and presence of `id`
- `result` precedence: `body` query param, then request `params`, then whole request

Rate limiting:
- Key: `protocol|method|normalizedPath|clientIP`
- Exceeded limit returns `429 Too Many Requests`

## OpenAPI (Spec-First)
OpenAPI should be maintained as an explicit artifact.
- Source of truth: `openapi/openapi.yaml` (to be created)
- Serve: `/openapi.yaml` and `/openapi.json`

Prefer updating the spec alongside behavior changes.

## Suggested Layout (Target)
Follow the design doc’s suggested package structure:
- `cmd/rudeserver/main.go`
- `internal/httpserver/router.go`
- `internal/scenario/parser.go`
- `internal/ip/ip.go`
- `internal/ratelimit/ratelimit.go`
- `internal/delay/delay.go`
- `internal/protocol/http.go`
- `internal/protocol/rest.go`
- `internal/protocol/jsonrpc.go`
- `internal/openapi/handler.go`
- `openapi/openapi.yaml`

## Engineering Guidelines
Implementation preferences:
- Use the Go standard library (`net/http`) unless there is a clear need
- Keep protocol adapters thin and reuse a shared middleware pipeline
- Avoid adding configuration systems for MVP; keep behavior per-request

Quality and verification:
- Prefer table-driven tests
- Use `net/http/httptest` for handler tests
- Run: `go test ./...` before claiming success (once the module exists)

## When In Doubt
Default to the design doc. If you need to change the design, update:
- `docs/plans/2026-01-27-http-test-service-design.md`
- `openapi/openapi.yaml` (when it exists)
