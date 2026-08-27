package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Bambi-uxx/Meowmico/backend/db"
)

func main() {
	db.Init()

	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Meowmico backend is alive! 🐱")
	})

	log.Printf("🐱 Meowmico backend starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
