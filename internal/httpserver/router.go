package httpserver

import (
	"net/http"

	"rudeserver/internal/delay"
	"rudeserver/internal/ip"
	"rudeserver/internal/protocol"
	"rudeserver/internal/ratelimit"
	"rudeserver/internal/scenario"
)

func NewRouter(store *ratelimit.Store) http.Handler {
	if store == nil {
		store = ratelimit.NewStore()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sc, err := scenario.ParseRequest(r)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		clientIP := ip.ClientIP(r)
		if !ratelimit.Allow(store, sc, clientIP) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch sc.Protocol {
			case scenario.ProtocolHTTP:
				protocol.WriteHTTP(w, sc)
			case scenario.ProtocolREST:
				protocol.WriteREST(w, sc)
			case scenario.ProtocolJSONRPC:
				protocol.HandleJSONRPC(w, r, sc)
			default:
				http.Error(w, "bad request", http.StatusBadRequest)
			}
		})

		delay.Wrap(handler, sc.Delay).ServeHTTP(w, r)
	})
}
