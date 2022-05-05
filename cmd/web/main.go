package main

import (
	"log"
	"net/http"
	"websockets-chat/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting channel listener goroutine...")
	go handlers.ListenToWsChannel()

	log.Println("Starting web server on port 8080...")
	_ = http.ListenAndServe(":8080", mux)
}
