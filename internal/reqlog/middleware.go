package reqlog

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"rudeserver/internal/ip"
)

const maxBodyBytes = 256 * 1024

type responseCapture struct {
	http.ResponseWriter
	status      int
	body        bytes.Buffer
	wroteHeader bool
}

func (r *responseCapture) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseCapture) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}

	if r.body.Len() < maxBodyBytes {
		remaining := maxBodyBytes - r.body.Len()
		if len(p) <= remaining {
			r.body.Write(p)
		} else {
			r.body.Write(p[:remaining])
		}
	}
	return r.ResponseWriter.Write(p)
}

func Middleware(store *Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()

		reqBytes, reqTrunc, reqSize, reqErr := readRequestBody(r)
		if reqBytes != nil {
			r.Body = io.NopCloser(bytes.NewReader(reqBytes))
		}

		capture := &responseCapture{ResponseWriter: w}
		next.ServeHTTP(capture, r)

		entry := Entry{
			Method:     r.Method,
			Path:       r.URL.Path,
			Query:      r.URL.RawQuery,
			Protocol:   protocolFromPath(r.URL.Path),
			ClientIP:   ip.ClientIP(r),
			UserAgent:  r.UserAgent(),
			Status:     capture.status,
			Duration:   time.Since(start).Milliseconds(),
			ReqHeaders: cloneHeaders(r.Header),
			ResHeaders: cloneHeaders(capture.Header()),
			ReqBody:    reqBytes,
			ResBody:    capture.body.Bytes(),
			ReqTruncated: reqTrunc,
			ResTruncated: capture.body.Len() >= maxBodyBytes,
			ReqSize:      reqSize,
			ResSize:      int64(capture.body.Len()),
			ContentType:  capture.Header().Get("Content-Type"),
			ReqError:     reqErr,
		}

		populateEncoding(&entry)
		store.Add(entry)
	})
}

func readRequestBody(r *http.Request) ([]byte, bool, int64, string) {
	if r.Body == nil {
		return nil, false, 0, ""
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes+1))
	if err != nil {
		return nil, false, 0, "read request body failed"
	}

	truncated := false
	if len(body) > maxBodyBytes {
		truncated = true
		body = body[:maxBodyBytes]
	}
	return body, truncated, int64(len(body)), ""
}

func protocolFromPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "http", "rest", "jsonrpc":
		return parts[0]
	default:
		return ""
	}
}

func skipPath(path string) bool {
	if path == "/" {
		return true
	}
	if strings.HasPrefix(path, "/ui/") {
		return true
	}
	if strings.HasPrefix(path, "/openapi.") {
		return true
	}
	return false
}

func cloneHeaders(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		cpy := make([]string, len(v))
		copy(cpy, v)
		out[k] = cpy
	}
	return out
}

func populateEncoding(entry *Entry) {
	if len(entry.ReqBody) > 0 {
		entry.ReqBodyIsUTF = utf8.Valid(entry.ReqBody)
		if !entry.ReqBodyIsUTF {
			entry.ReqBodyB64 = base64.StdEncoding.EncodeToString(entry.ReqBody)
		}
	}
	if len(entry.ResBody) > 0 {
		entry.ResBodyIsUTF = utf8.Valid(entry.ResBody)
		if !entry.ResBodyIsUTF {
			entry.ResBodyB64 = base64.StdEncoding.EncodeToString(entry.ResBody)
		}
	}
}
