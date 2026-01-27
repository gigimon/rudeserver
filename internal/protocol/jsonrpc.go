package protocol

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"rudeserver/internal/scenario"
)

func HandleJSONRPC(w http.ResponseWriter, r *http.Request, sc scenario.Scenario) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var request map[string]any
	if err := json.Unmarshal(bodyBytes, &request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if request["jsonrpc"] != "2.0" {
		http.Error(w, "invalid jsonrpc", http.StatusBadRequest)
		return
	}

	id, ok := request["id"]
	if !ok {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	result := resolveResult(sc.Body, request)

	response := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}

	writeHeaders(w, sc.Headers)
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}

	status := sc.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func resolveResult(body string, request map[string]any) any {
	if body != "" {
		trimmed := strings.TrimSpace(body)
		if trimmed != "" {
			var parsed any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
				return parsed
			}
		}
		return body
	}
	if params, ok := request["params"]; ok {
		return params
	}
	return request
}
