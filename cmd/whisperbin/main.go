package main

import (
	"log"
	"net/http"
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

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Server running at http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
