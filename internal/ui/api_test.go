package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"rudeserver/internal/reqlog"
)

func TestAPIListAndDetail(t *testing.T) {
	store := reqlog.NewStore(10)
	store.Add(reqlog.Entry{Method: "GET", Path: "/http/status/200", Status: 200})
	store.Add(reqlog.Entry{Method: "POST", Path: "/rest/status/201", Status: 201})

	h := APIHandler(store)

	// list
	listReq := httptest.NewRequest(http.MethodGet, "/ui/api/requests", nil)
	listRec := httptest.NewRecorder()
	h.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRec.Code)
	}
	var list []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &list); err != nil {
		t.Fatalf("list json: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d", len(list))
	}
	idAny := list[0]["id"]
	idFloat, ok := idAny.(float64)
	if !ok {
		t.Fatalf("id type = %T", idAny)
	}
	id := int64(idFloat)

	// detail
	detailReq := httptest.NewRequest(http.MethodGet, "/ui/api/requests/"+formatID(id), nil)
	detailRec := httptest.NewRecorder()
	h.ServeHTTP(detailRec, detailReq)

	if detailRec.Code != http.StatusOK {
		t.Fatalf("detail status = %d", detailRec.Code)
	}
	var detail map[string]any
	if err := json.Unmarshal(detailRec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("detail json: %v", err)
	}
	if detail["id"].(float64) != float64(id) {
		t.Fatalf("detail id mismatch")
	}
}

func TestAPIMissingID(t *testing.T) {
	store := reqlog.NewStore(10)
	h := APIHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/ui/api/requests/999", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func formatID(id int64) string {
	return fmt.Sprintf("%d", id)
}
