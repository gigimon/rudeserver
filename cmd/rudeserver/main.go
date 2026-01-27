package main

import (
	"log"
	"net/http"
	"time"

	"rudeserver/internal/openapi"
)

func main() {
	mux := http.NewServeMux()

	if err := openapi.Register(mux); err != nil {
		log.Fatalf("openapi setup error: %v", err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("rudeserver"))
	})

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
