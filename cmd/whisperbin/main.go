package main

import (
	"log"
	"net/http"

	"whisperbin/internal/storage"
	"whisperbin/internal/web"
)

func main() {
	store := storage.NewStore()
	handler := web.NewHandler(store)

	http.Handle("/", handler.Routes())
	http.HandleFunc("/ws", handler.WebSocketHandler)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
