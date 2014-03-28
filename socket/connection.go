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

// connection is a middleman between the websocket connection and redis.
type S struct {
	name string
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Buffered channel of inbound messages.
	receive chan []byte

	// Buffered channel of inbound messages.
	onClose func()
}

func NewSocket(name string, w http.ResponseWriter, r *http.Request) (*S, error) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
		return nil, err
	} else if err != nil {
		http.Error(w, "Error while opening websocket!", http.StatusInternalServerError)
		return nil, err
	}

	s := &S{
		name: name,
		ws: ws,
		send: make(chan []byte, 256),
	}

	go s.writePump()
	go s.readPump()
	return s, nil
}

func (s *S) SetOnCloseFunc(oc func()) {
	s.onClose = oc
}

func (s *S) Read() []byte {
	msg := <- s.receive
	return msg
}

func (s *S) Write(payload []byte) {
	s.send <- payload
}

func (s *S) readPump() {
	defer func() {
		if s.onClose != nil {
			s.onClose()
		}
		s.ws.Close()
	}()
	for {
		_, message, err := s.ws.ReadMessage()
		if err != nil {
			return
		}

		s.receive <- message
	}
}

// writePump pumps messages to the websocket connection.
func (s *S) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if s.onClose != nil {
			s.onClose()
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
func (s *S) write(mt int, payload []byte) error {
	s.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return s.ws.WriteMessage(mt, payload)
}
