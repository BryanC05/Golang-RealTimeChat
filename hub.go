package main

import (
	"encoding/json"
	"log"
	"time"
)

type Message struct {
	Type      string `json:"type"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

type Hub struct {
	clients map[*Client]bool

	broadcast chan *Message

	register chan *Client

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

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client %s registered", client.username)

			joinMsg := &Message{
				Type:      "join",
				Username:  client.username,
				Timestamp: time.Now().Format(time.Kitchen),
			}
			h.broadcastMessage(joinMsg)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client %s unregistered", client.username)

				leaveMsg := &Message{
					Type:      "leave",
					Username:  client.username,
					Timestamp: time.Now().Format(time.Kitchen),
				}
				h.broadcastMessage(leaveMsg)
			}

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) broadcastMessage(message *Message) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message: %v", err)
		return
	}

	for client := range h.clients {
		select {
		case client.send <- msgBytes:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}

}
