package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

var hub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *Hub) register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
	log.Println("New client connected! Total:", len(h.clients))
}

func (h *Hub) unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
	log.Println("Client disconnected. Total:", len(h.clients))
}

func (h *Hub) broadcast(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Error broadcasting:", err)
			delete(h.clients, conn)
			conn.Close()
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	hub.register(conn)
	defer hub.unregister(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
