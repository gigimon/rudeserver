# UI Session Counter Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a session-wide request counter to the UI that shows total requests since start.

**Architecture:** Extend the in-memory request log store with a monotonically increasing `total` counter, expose it via the `/ui/api/requests` response, and render it in the UI stats.

**Tech Stack:** Go (`net/http`, `httptest`), vanilla HTML/CSS/JS.

### Task 1: Store Total Counter

**Files:**
- Modify: `internal/reqlog/store.go`
- Test: `internal/reqlog/store_test.go`

**Step 1: Write the failing test**

```go
func TestStoreTotalCount(t *testing.T) {
    store := NewStore(2)
    store.Add(Entry{})
    store.Add(Entry{})
    store.Add(Entry{})
    if store.Total() != 3 {
        t.Fatalf("total = %d", store.Total())
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/reqlog -v`  
Expected: FAIL with “Total undefined”

**Step 3: Write minimal implementation**

- Add `total` field to `Store`  
- Increment it on every `Add`  
- Add `Total() int64`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/reqlog -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/reqlog/store.go internal/reqlog/store_test.go
git commit -m "feat: track total request count"
```

### Task 2: API Payload + UI Rendering

**Files:**
- Modify: `internal/ui/api.go`
- Modify: `internal/ui/api_test.go`
- Modify: `internal/ui/assets/app.js`
- Modify: `internal/ui/assets/index.html`

**Step 1: Write the failing test**

Update API test to expect:
- `GET /ui/api/requests` returns `{ items: [...], total: N }`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/ui -v`  
Expected: FAIL (payload shape mismatch)

**Step 3: Write minimal implementation**

- Include `total` in list response  
- Use `items` for list  
- Render `total` in UI stats (e.g. “Session Total”)

**Step 4: Run test to verify it passes**

Run: `go test ./internal/ui -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/ui/api.go internal/ui/api_test.go internal/ui/assets/app.js internal/ui/assets/index.html
git commit -m "feat: show session total in ui"
```

### Task 3: Full Verification

**Step 1: Run tests**

Run: `go test ./...`  
Expected: PASS

