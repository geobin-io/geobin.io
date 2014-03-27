package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 60 * time.Second
)

// connection is a middleman between the websocket connection and redis.
type connection struct {
	name string
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// writePump pumps messages from redis to the websocket connection.
func (c *connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := pubsub.Unsubscribe(c.name); err != nil {
			log.Println("Failure to UNSUBSCRIBE from", c.name, err)
		}
		delete(sockets, c.name)
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// writes a message with the given message type and payload.
func (c *connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func openSocket(w http.ResponseWriter, r *http.Request) {
	// upgrade the connection
	name := mux.Vars(r)["name"]

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
		return
	} else if err != nil {
		log.Println("Error upgrading connection to websocket protocol:", err)
		http.Error(w, "Error while opening websocket!", http.StatusInternalServerError)
		return
	}

	// start pub subbing
	if err = pubsub.Subscribe(name); err != nil {
		log.Println("Failure to SUBSCRIBE to", name, err)
	}

	c := &connection{
		name,
		ws,
		make(chan []byte, 256),
	}

	// keep track of the outbound channel for pubsubbery
	sockets[name] = c.send

	// this routine keeps the ws open with pings, and checks for/sends outbound messages
	go c.writePump()
}
