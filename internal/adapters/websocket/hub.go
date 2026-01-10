package websocket

import (
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients.
	// Map of specific users to their active connections.
	// A user might have multiple connections (e.g., phone + laptop), so we store a list or map.
	// For simplicity, we'll assume one connection per user for now, or map Key=UserID, Value=Client
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	mu sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// We map the client by their UserID (ShopperID or CustomerID)
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("[WS] Client registered: %s", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] Client unregistered: %s", client.UserID)

		case message := <-h.broadcast:
			// Global broadcast (rarely used in this app, but good to have)
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.UserID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to a specific user if they are connected.
func (h *Hub) SendToUser(userID string, message interface{}) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		client.SendJSON(message)
	} else {
		// log.Printf("[WS] User %s not connected, message dropped (or push notif)", userID)
	}
}

func (h *Hub) Broadcast(message interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		client.SendJSON(message)
	}
}
