package main

import (
	"log"
	"net/http"
	"time"

	"rudeserver/internal/httpserver"
	"rudeserver/internal/openapi"
	"rudeserver/internal/ratelimit"
)

func main() {
	mux := http.NewServeMux()

	if err := openapi.Register(mux); err != nil {
		log.Fatalf("openapi setup error: %v", err)
	}

	mux.Handle("/", httpserver.NewRouter(ratelimit.NewStore()))

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
