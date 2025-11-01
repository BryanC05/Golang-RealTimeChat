package main

import (
	"log"
	"net/http"
)

// serveHome serves the homepage HTML file.
func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

// serveWs handles the WebSocket connection request.
// It gets the username from the query parameter (e.g., /ws?username=Alex).
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// Get username from query param
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Upgrade the HTTP connection to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create a new client
	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256), // Buffered channel
		username: username,
	}

	// Register the client with the hub
	client.hub.register <- client

	// Start the client's read and write goroutines
	go client.writePump()
	go client.readPump()
}

func main() {
	// Create a new hub (chat room)
	hub := newHub()
	// Run the hub in its own goroutine
	go hub.run()

	// Define the HTTP routes
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Println("Server starting on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}