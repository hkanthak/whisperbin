package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"whisperbin/internal"
	"whisperbin/internal/storage"
	"whisperbin/internal/web"
)

func main() {
	store := storage.NewStore()
	handler := web.NewHandler(store)

	go func() {
		ticker := time.NewTicker(internal.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			store.CleanupExpired()
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("Server running at http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}
