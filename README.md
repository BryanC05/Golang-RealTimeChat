# Go Real-Time Chat Application

This is a simple, real-time chat server built with Go (Golang). It uses **WebSockets** for persistent, bidirectional communication and **Goroutines/Channels** to safely manage multiple client connections and broadcast messages concurrently.

The project includes a minimal HTML/CSS/JS frontend to act as the chat client.

-----

## ✨ Features

  * **Real-Time Communication:** Instantly broadcast messages to all connected clients.
  * **Concurrency-Safe:** Uses a central "Hub" with channels to prevent race conditions when managing clients.
  * **Client Notifications:** Automatically notifies the room when a user joins or leaves.
  * **Simple UI:** Includes a basic `index.html` to demonstrate functionality.

-----

## 🛠️ Tech Stack

  * **Backend:** Go (Golang)
  * **WebSockets:** `github.com/gorilla/websocket`
  * **Frontend:** Vanilla HTML, CSS, and JavaScript

-----

## 🚀 How to Run

### Prerequisites

  * Go (Version 1.18 or newer)

### 1\. Get Dependencies

From your project's root folder, initialize the Go module and get the `gorilla/websocket` package:

```bash
go mod init go-chat
go get github.com/gorilla/websocket
```

### 2\. Run the Server

Make sure all four files (`main.go`, `hub.go`, `client.go`, `index.html`) are in the same directory. Then, run:

```bash
go run .
```

The server will start on `http://localhost:8080`.

### 3\. Test the Chat

1.  Open your browser and go to **`http://localhost:8080/`**.
2.  Enter a username and click "Join".
3.  Open a **second browser window** (or a new incognito window) and go to the same address.
4.  Enter a different username and click "Join".
5.  You can now send messages between the two windows in real-time.

-----

## 📡 API Endpoints

  * **`GET /`**

      * **Description:** Serves the `index.html` file, which contains the chat UI.

  * **`GET /ws?username={name}`**

      * **Description:** This is the main WebSocket endpoint. The "login" is handled by passing the username as a query parameter. This endpoint upgrades the client's HTTP connection to a persistent WebSocket connection and registers them with the chat hub.

-----

## 🧠 Core Concepts

This project is a classic example of concurrent programming in Go. It works by coordinating multiple goroutines using channels.

### 1\. The Hub (`hub.go`)

The **Hub** is the central chat room. It runs in a **single, dedicated goroutine** (started in `main.go`). It is responsible for:

  * Keeping track of all active clients.
  * Registering new clients.
  * Unregistering disconnected clients.
  * Broadcasting messages to all clients.

It uses channels (`register`, `unregister`, `broadcast`) to receive events. By handling all these events in one goroutine using a `select` statement, we **prevent race conditions**. No two operations can access the client list at the same time, so we don't need to use mutexes (locks).

### 2\. The Client (`client.go`)

Each user who connects to the server is represented by a **Client** object. Every `Client` runs **two dedicated goroutines**:

1.  **`readPump()` (Inbound):** This goroutine runs in a loop, reading new messages from the client's WebSocket connection. When it receives a message, it sends it to the **Hub's `broadcast` channel**.
2.  **`writePump()` (Outbound):** This goroutine runs in a loop, waiting for messages to arrive on the **Client's personal `send` channel**. When the Hub broadcasts a message, it sends it to this channel. The `writePump` then writes that message to the client's WebSocket connection.

### 3\. Concurrency Flow

The flow of a single message looks like this:

1.  **User A** types a message and hits Send.
2.  **User A's** `readPump()` goroutine reads the message from the WebSocket.
3.  `readPump()` sends the message to the **Hub's** `broadcast` channel.
4.  The **Hub's** `run()` goroutine (in its `select` loop) receives the message.
5.  The **Hub** loops over every registered client and sends the message to their individual `send` channels.
6.  **User A's** `writePump()` receives the message on its `send` channel and writes it to their WebSocket (so they see their own message).
7.  **User B's** `writePump()` receives the *same* message on its `send` channel and writes it to their WebSocket.

-----

## ⚠️ Note on Persistence

This is an **ephemeral** chat application. Chat history is **not** stored in any database. All messages are held in memory only long enough to be broadcast. If the server restarts, all chat history is lost.
