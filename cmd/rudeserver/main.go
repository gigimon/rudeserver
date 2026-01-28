package main

import (
	"log"
	"net/http"
	"time"

	"rudeserver/internal/httpserver"
	"rudeserver/internal/openapi"
	"rudeserver/internal/ratelimit"
	"rudeserver/internal/reqlog"
	"rudeserver/internal/ui"
)

func main() {
	mux := http.NewServeMux()

	if err := openapi.Register(mux); err != nil {
		log.Fatalf("openapi setup error: %v", err)
	}

	uiHandler, err := ui.Handler()
	if err != nil {
		log.Fatalf("ui setup error: %v", err)
	}
	mux.Handle("/", uiHandler)

	logStore := reqlog.NewStore(100)
	apiHandler := httpserver.NewRouter(ratelimit.NewStore())
	loggedAPI := reqlog.Middleware(logStore, apiHandler)

	mux.Handle("/ui/api/", ui.APIHandler(logStore))
	mux.Handle("/http/", loggedAPI)
	mux.Handle("/rest/", loggedAPI)
	mux.Handle("/jsonrpc/", loggedAPI)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("rudeserver listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
