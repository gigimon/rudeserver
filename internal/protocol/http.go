package protocol

import (
	"net/http"

	"rudeserver/internal/scenario"
)

func WriteHTTP(w http.ResponseWriter, sc scenario.Scenario) {
	writeHeaders(w, sc.Headers)
	status := sc.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	if sc.Body != "" {
		_, _ = w.Write([]byte(sc.Body))
	}
}

func writeHeaders(w http.ResponseWriter, headers http.Header) {
	for name, values := range headers {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
}

