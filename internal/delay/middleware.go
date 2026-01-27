package delay

import (
	"net/http"
	"time"
)

// Wrap sleeps for the given delay before invoking the next handler.
func Wrap(next http.Handler, delay time.Duration) http.Handler {
	if delay <= 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		next.ServeHTTP(w, r)
	})
}

