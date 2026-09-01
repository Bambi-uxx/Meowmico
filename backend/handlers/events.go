package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bambi-uxx/Meowmico/backend/db"
)

var BroadcastFunc func([]byte)

type Event struct {
	ID        int       `json:"id"`
	Channel   string    `json:"channel"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(
		"INSERT INTO events (channel, content) VALUES (?, ?)",
		event.Channel, event.Content,
	)
	if err != nil {
		http.Error(w, "Failed to save event", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	event.ID = int(id)
	event.CreatedAt = time.Now()

	// Broadcast to WebSocket clients AFTER saving
	if BroadcastFunc != nil {
		data, _ := json.Marshal(event)
		BroadcastFunc(data)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func GetEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(
		"SELECT id, channel, content, created_at FROM events ORDER BY created_at DESC LIMIT 20",
	)
	if err != nil {
		http.Error(w, "Failed to fetch events", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Channel, &e.Content, &e.CreatedAt); err != nil {
			continue
		}
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error reading events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
