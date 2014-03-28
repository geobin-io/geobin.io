// The `socket` package wraps the `github.com/gorilla/websocket` package with a basic implementation
// based on their sample code, with the goal of greatly reducing the interface for some generalized
// use cases.
package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 60 * time.Second
)

// S exposes some methods for interacting with a websocket
type S interface {
	// Submits a payload to the web socket as a text message.
	Write([]byte)
	// Set the func that's called when a message is read from the socket.
	// The call is made from a separate routine.
	// The message types are defined in RFC 6455, section 11.8.
	SetOnRead(func(int, []byte))
	// Set the func that's called when the socket is being closed, typically due to a timeout
	// when sending a message or possibly another io error in either direction.
	// The call is made from a separate routine.
	SetOnClose(func(string))
}

// implementation of S
type s struct {
	// a string associated with the socket
	name string

	// the websocket connection
	ws *websocket.Conn

	// buffered channel of outbound messages
	send chan []byte

	// event functions
	onReceive func(messageType int, message []byte)
	onClose func(name string)
}

// Upgrade an existing TCP connection to a websocket connection in response to a client request for a websocket.
// `name` here is just an identifying string for the socket, which will be returned when/if the socket is closed
// in a call to the that can be reference with `SetOnClose()`.
func NewSocket(name string, w http.ResponseWriter, r *http.Request) (S, error) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
		return nil, err
	} else if err != nil {
		http.Error(w, "Error while opening websocket!", http.StatusInternalServerError)
		return nil, err
	}

	s := &s{
		name: name,
		ws: ws,
		send: make(chan []byte, 256),
	}

	go s.writePump()
	go s.readPump()
	return s, nil
}


func (s *s) SetOnClose(oc func(name string)) {
	s.onClose = oc
}

func (s *s) SetOnRead(or func(messageType int, message []byte)) {
	s.onReceive = or
}

func (s *s) Write(payload []byte) {
	s.send <- payload
}

// readPump pumps messages from the websocket
func (s *s) readPump() {
	defer func() {
		if s.onClose != nil {
			s.onClose(s.name)
		}
		s.ws.Close()
	}()
	for {
		mt, message, err := s.ws.ReadMessage()
		if err != nil {
			return
		}

		if s.onReceive != nil {
			s.onReceive(mt, message)
		}
	}
}

// writePump pumps messages to the websocket
func (s *s) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if s.onClose != nil {
			s.onClose(s.name)
		}
		s.ws.Close()
	}()
	for {
		select {
		case message := <-s.send:
			if err := s.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := s.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// writes a message with the given message type and payload.
func (s *s) write(mt int, payload []byte) error {
	s.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return s.ws.WriteMessage(mt, payload)
}
