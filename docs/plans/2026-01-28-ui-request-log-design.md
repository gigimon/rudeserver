# UI Request Log (In‑Memory) — Design

**Goal:** Provide a lightweight UI at `/` that shows the latest 100 requests and their responses, with detailed views in both “pretty” and raw formats.  
**Data durability:** in‑memory only; data is lost on restart.

## Scope
- Track all requests **except** `/` (UI), `/ui/*` (static assets + API), and `/openapi.*`.
- Store up to **100** entries in a fixed ring buffer.
- Cap request and response bodies at **256 KB** each.

## UI & UX
- UI served at `/` with static assets under `/ui/*`.
- Light “inspector” aesthetic: card layout, gentle shadows, clear status coloring.
- Left panel: list of recent requests (time, method, path, status, duration, client IP).
- Right panel: detail view with tabs **Overview**, **Request**, **Response**, **Raw**.
- Pretty view attempts JSON formatting; fallback to text.  
- Raw view shows exact stored bytes; if not UTF‑8, show base64 with copy button.
- Polling: `GET /ui/api/requests` every 3–5 seconds.

## API Endpoints
- `GET /ui/api/requests` → array of summaries.
- `GET /ui/api/requests/{id}` → full request/response record.

## Logging Middleware
Wrap the main router with a middleware that:
1. Reads the request body up to 256 KB and **replaces** `r.Body` so handlers can read it normally.
2. Wraps `http.ResponseWriter` to capture status, headers, and response body (up to 256 KB).
3. Stores an entry in a ring buffer with a monotonically increasing `id`.

## Data Model (core fields)
`id`, `timestamp`, `method`, `path`, `query`, `protocol`, `client_ip`, `status`, `duration_ms`,  
`req_headers`, `res_headers`, `req_body`, `res_body`, `req_truncated`, `res_truncated`, `content_type`.

## Error Handling
- If a requested `id` is missing (evicted), return `404`.
- Truncated bodies are flagged and include original size.
- Minimal `500` responses for unexpected serialization issues.

## Testing (no UI smoke test)
- Ring buffer: eviction after 100, monotonic IDs.
- Middleware: body capture + passthrough, response capture, truncation.
- Exclusions: `/`, `/ui/*`, `/openapi.*` not logged.
- API: list and detail endpoints, missing ID → 404.
