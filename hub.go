package main

import ()

type Hub struct {
	// Channel to register new connections
	Register chan *Connection

	// Channel to unregister new connections
	Unregister chan *Connection

	// Channel to send broadcast messages
	Broadcast chan string

	// Connections storage
	connections map[*Connection]bool
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.Register:
			h.connections[c] = true
		case c := <-h.Unregister:
			delete(h.connections, c)
		case msg := <-h.Broadcast:
			h.broadcast(msg)
		}
	}
}

func (h *Hub) broadcast(msg string) {
	for connection, _ := range h.connections {
		connection.inputChannel <- msg
	}
}
