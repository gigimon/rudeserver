# RudeServer

RudeServer is a small Go HTTP test service for exercising HTTP clients under controlled conditions.

**Features**
- Any HTTP status code
- Per-(protocol + method + path + client IP) rate limiting
- Optional response delay
- Protocol adapters via URL prefixes: `/http`, `/rest`, `/jsonrpc`
- Spec-first OpenAPI endpoints

## Quickstart

```bash
go run ./cmd/rudeserver
```

Server listens on `:8080`.

## Docker

```bash
docker build -t rudeserver .
docker run --rm -p 8080:8080 rudeserver
```

## OpenAPI

- `GET /openapi.yaml`
- `GET /openapi.json`

## URL Shape

Paths:
- `/http/status/{code}`
- `/rest/status/{code}`
- `/jsonrpc/status/{code}` (POST only)

Query parameters (shared):
- `rl`: rate limit (RPS)
- `burst`: burst size (requires `rl`)
- `delay`: response delay (Go duration, e.g. `200ms`, `1s`)
- `body`: response body (string)
- `h`: response header, repeatable, `Name:Value`

## Examples

### Basic HTTP status
```bash
curl -i http://localhost:8080/http/status/418
```

### Body + headers
```bash
curl -i "http://localhost:8080/http/status/200?body=hello&h=Content-Type:text/plain&h=X-Test:1"
```

### Rate limit per client
```bash
curl -i "http://localhost:8080/http/status/200?rl=1&burst=1"
curl -i "http://localhost:8080/http/status/200?rl=1&burst=1"  # 429
```

### Delay
```bash
curl -i "http://localhost:8080/rest/status/200?delay=250ms"
```

### JSON-RPC (POST only)
```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"params":{"a":1}}' \
  "http://localhost:8080/jsonrpc/status/200"
```

### JSON-RPC custom result
```bash
curl -i \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"params":{"a":1}}' \
  "http://localhost:8080/jsonrpc/status/200?body={\"ok\":true}"
```

## Notes

- `/http` and `/rest` accept any HTTP method.
- `/jsonrpc` validates `jsonrpc: "2.0"` and requires `id`.
