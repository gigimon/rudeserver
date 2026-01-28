package ui

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"rudeserver/internal/reqlog"
)

func APIHandler(store *reqlog.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/ui/api")
		switch {
		case path == "/requests":
			writeJSON(w, toSummaries(store.List()))
			return
		case strings.HasPrefix(path, "/requests/"):
			idStr := strings.TrimPrefix(path, "/requests/")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			entry, ok := store.Find(id)
			if !ok {
				http.NotFound(w, r)
				return
			}
			writeJSON(w, toDetail(entry))
			return
		default:
			http.NotFound(w, r)
			return
		}
	})
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(payload)
}

func toSummaries(entries []reqlog.Entry) []map[string]any {
	out := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]any{
			"id":          e.ID,
			"timestamp":   e.Timestamp,
			"method":      e.Method,
			"path":        e.Path,
			"query":       e.Query,
			"protocol":    e.Protocol,
			"client_ip":   e.ClientIP,
			"status":      e.Status,
			"duration_ms": e.Duration,
			"content_type": e.ContentType,
		})
	}
	return out
}

func toDetail(e reqlog.Entry) map[string]any {
	return map[string]any{
		"id":           e.ID,
		"timestamp":    e.Timestamp,
		"method":       e.Method,
		"path":         e.Path,
		"query":        e.Query,
		"protocol":     e.Protocol,
		"client_ip":    e.ClientIP,
		"status":       e.Status,
		"duration_ms":  e.Duration,
		"user_agent":   e.UserAgent,
		"req_headers":  e.ReqHeaders,
		"res_headers":  e.ResHeaders,
		"req_body":     string(e.ReqBody),
		"res_body":     string(e.ResBody),
		"req_trunc":    e.ReqTruncated,
		"res_trunc":    e.ResTruncated,
		"req_size":     e.ReqSize,
		"res_size":     e.ResSize,
		"req_utf8":     e.ReqBodyIsUTF,
		"res_utf8":     e.ResBodyIsUTF,
		"req_b64":      e.ReqBodyB64,
		"res_b64":      e.ResBodyB64,
		"content_type": e.ContentType,
		"req_error":    e.ReqError,
		"res_error":    e.ResError,
	}
}
