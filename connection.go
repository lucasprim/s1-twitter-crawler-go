package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Connection struct {
	// Channel to send messages to this connection
	inputChannel chan string

	// Websocket Conn
	ws *websocket.Conn
}

func (c *Connection) StartWriter() {
	// Ping messages ticker
	ticker := time.NewTicker(58 * time.Second)

	defer func() {
		hub.Unregister <- c
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case s := <-c.inputChannel:
			if err := c.ws.WriteMessage(websocket.TextMessage, []byte(s)); err != nil {
				return
			}
		}
	}
}

func (c *Connection) StartReader() {
	defer func() {
		hub.Unregister <- c
		c.ws.Close()
	}()

	c.ws.SetReadLimit(512)
	c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.ws.ReadMessage()

		if err != nil {
			fmt.Println("ERROR! ", err)
			break
		}
	}
}
