package main

import (
	"log"
	"net/http"
	"time"

	"whisperbin/internal/storage"
	"whisperbin/internal/web"
)

func main() {
	store := storage.NewStore()
	handler := web.NewHandler(store)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			store.CleanupExpired()
		}
	}()

	http.Handle("/", handler.Routes())
	http.HandleFunc("/ws", handler.WebSocketHandler)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
