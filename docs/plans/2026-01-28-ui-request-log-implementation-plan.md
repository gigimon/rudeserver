# UI Request Log Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a lightweight UI at `/` that shows the latest 100 requests with full request/response details, stored in memory and polled via `/ui/api/*`.

**Architecture:** Implement a request/response logging middleware that captures metadata and capped bodies into a fixed-size ring buffer. Expose JSON APIs under `/ui/api/*` and serve a static HTML/CSS/JS UI at `/` with assets under `/ui/*`.

**Tech Stack:** Go (`net/http`, `httptest`, `embed`), vanilla HTML/CSS/JS.

### Task 1: Ring Buffer Store (In‑Memory)

**Files:**
- Create: `internal/reqlog/store.go`
- Test: `internal/reqlog/store_test.go`

**Step 1: Write the failing test**

```go
func TestStoreEvictsOldest(t *testing.T) {
    store := NewStore(3)
    store.Add(Entry{ID: 1})
    store.Add(Entry{ID: 2})
    store.Add(Entry{ID: 3})
    store.Add(Entry{ID: 4})

    got := store.List()
    if len(got) != 3 || got[0].ID != 4 || got[2].ID != 2 {
        t.Fatalf("unexpected order or eviction: %+v", got)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/reqlog -v`  
Expected: FAIL with “NewStore undefined” (or similar)

**Step 3: Write minimal implementation**

- Implement `Entry`, `Store`, `NewStore(max int)`  
- `Add(entry)` assigns monotonically increasing `ID`  
- `List()` returns newest-first slice (max `N`)

**Step 4: Run test to verify it passes**

Run: `go test ./internal/reqlog -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/reqlog/store.go internal/reqlog/store_test.go
git commit -m "feat: add in-memory request log store"
```

### Task 2: Request/Response Logging Middleware

**Files:**
- Create: `internal/reqlog/middleware.go`
- Modify: `internal/reqlog/store.go`
- Test: `internal/reqlog/middleware_test.go`

**Step 1: Write the failing test**

Test that:
- Request body is captured and still readable downstream
- Response body + status are captured
- Bodies are capped at 256 KB
- Excludes `/`, `/ui/*`, `/openapi.*`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/reqlog -v`  
Expected: FAIL (missing middleware)

**Step 3: Write minimal implementation**

- Wrap `http.ResponseWriter` to capture status + body (cap to 256 KB)  
- For request body: capture via tee reader or read+restore (cap to 256 KB)  
- Store entries in `Store` with timestamps, method, path, query, protocol, status, duration, headers, body, truncated flags  
- Skip logging for `/`, `/ui/`, `/openapi`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/reqlog -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/reqlog/middleware.go internal/reqlog/middleware_test.go internal/reqlog/store.go
git commit -m "feat: add request/response logging middleware"
```

### Task 3: UI API Endpoints

**Files:**
- Create: `internal/ui/api.go`
- Test: `internal/ui/api_test.go`

**Step 1: Write the failing test**

Test:
- `GET /ui/api/requests` returns JSON list
- `GET /ui/api/requests/{id}` returns detail
- Missing ID returns 404

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui -v`  
Expected: FAIL (missing handlers)

**Step 3: Write minimal implementation**

- JSON encoding of summary and detail records  
- Route parsing for `/ui/api/requests` and `/ui/api/requests/{id}`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/api.go internal/ui/api_test.go
git commit -m "feat: add ui api endpoints"
```

### Task 4: UI Static Assets + Handler

**Files:**
- Create: `internal/ui/handler.go`
- Create: `internal/ui/assets/index.html`
- Create: `internal/ui/assets/app.js`
- Create: `internal/ui/assets/styles.css`

**Step 1: Write the failing test**

No UI smoke test (per requirement). Skip test creation.

**Step 2: Implement static handler**

- Serve `/` → `index.html`  
- Serve `/ui/*` → embedded assets  
- Use `embed` and `fs.Sub`  

**Step 3: Commit**

```bash
git add internal/ui/handler.go internal/ui/assets/index.html internal/ui/assets/app.js internal/ui/assets/styles.css
git commit -m "feat: add ui static assets"
```

### Task 5: Wire Everything in `main`

**Files:**
- Modify: `cmd/rudeserver/main.go`
- Modify: `internal/httpserver/router.go`
- Test: `internal/httpserver/router_test.go` (if needed)

**Step 1: Update wiring**

- Create `reqlog.Store` with max 100  
- Wrap API router with logging middleware  
- Register UI handler at `/` and `/ui/*`  
- Register UI API under `/ui/api/*`  
- Ensure `/http/*`, `/rest/*`, `/jsonrpc/*` route to API handler

**Step 2: Run full tests**

Run: `go test ./...`  
Expected: PASS

**Step 3: Commit**

```bash
git add cmd/rudeserver/main.go internal/httpserver/router.go internal/httpserver/router_test.go
git commit -m "feat: wire request log ui"
```

### Task 6: Final Verification

**Step 1: Run tests**

Run: `go test ./...`  
Expected: PASS

**Step 2: Commit any last doc tweaks**

Only if needed.
