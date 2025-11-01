package main

import (
	"encoding/json"
	"log"
	"time"
)

// Message struct defines the JSON message format
type Message struct {
	Type      string `json:"type"`      // "chat", "join", "leave"
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients. A map is used for O(1) lookups.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// run is the main loop for the hub. It must be run as a goroutine.
func (h *Hub) run() {
	for {
		// The select statement blocks until one of its cases is ready.
		select {
		case client := <-h.register:
			// A new client has connected.
			h.clients[client] = true
			log.Printf("Client %s registered", client.username)

			// Create a "join" message and broadcast it.
			joinMsg := &Message{
				Type:      "join",
				Username:  client.username,
				Timestamp: time.Now().Format(time.Kitchen),
			}
			h.broadcastMessage(joinMsg)

		case client := <-h.unregister:
			// A client has disconnected.
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client %s unregistered", client.username)

				// Create a "leave" message and broadcast it.
				leaveMsg := &Message{
					Type:      "leave",
					Username:  client.username,
					Timestamp: time.Now().Format(time.Kitchen),
				}
				h.broadcastMessage(leaveMsg)
			}

		case message := <-h.broadcast:
			// A client has sent a chat message.
			h.broadcastMessage(message)
		}
	}
}

// broadcastMessage marshals the message to JSON and sends it to all clients.
func (h *Hub) broadcastMessage(message *Message) {
	// Marshal the message to JSON once.
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	// Send the message to all connected clients.
	for client := range h.clients {
		select {
		case client.send <- msgBytes:
			// Message sent
		default:
			// Send buffer is full; assume client is dead or slow.
			close(client.send)
			delete(h.clients, client)
		}
	}
}