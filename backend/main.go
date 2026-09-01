package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Bambi-uxx/Meowmico/backend/db"
	"github.com/Bambi-uxx/Meowmico/backend/handlers"
)

func main() {
	db.Init()

	handlers.BroadcastFunc = hub.broadcast

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Meowmico backend is alive!")
	})
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateEvent(w, r)
		} else {
			handlers.GetEvents(w, r)
		}
	})
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateMessage(w, r)
		} else {
			handlers.GetMessages(w, r)
		}
	})
	http.HandleFunc("/ws", wsHandler)

	log.Printf("Meowmico backend starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
