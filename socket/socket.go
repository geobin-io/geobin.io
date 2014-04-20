// The `socket` package wraps the `github.com/gorilla/websocket` package with a basic implementation
// based on their sample code, with the goal of greatly reducing the interface for some generalized
// use cases.
package socket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"log"
	"net/url"
	"net"
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
	// return the provided name
	GetName() string
	// Close the socket.
	Close()
}

// implementation of S
type s struct {
	// a string associated with the socket
	name string

	// the websocket connection
	ws *websocket.Conn

	// buffered channel of outbound messages
	send chan []byte

	shutdown chan bool

	// event functions
	onRead func(messageType int, message []byte)
	onClose func(name string)
}

// Upgrade an existing TCP connection to a websocket connection in response to a client request for a websocket.
// `name` here is just an identifying string for the socket, which will be returned when/if the socket is closed
// by calling a provided function (settable with `SetOnClose()`).
// `or` here is the func that's called when a message is read from the socket. The call is made from a separate routine.
// The message types are defined in RFC 6455, section 11.8.
// `oc` here is the func that's called when the socket is just about to be closed. The call is made from a
// separate routine.
// If you do not care about these callbacks, pass nil instead.
func NewSocket(name string, w http.ResponseWriter, r *http.Request, or func(int, []byte), oc func(string)) (S, error) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
		return nil, err
	} else if err != nil {
		http.Error(w, "Error while opening websocket!", http.StatusInternalServerError)
		return nil, err
	}

	return socketSetup(name, ws, or, oc), nil
}

// Create a client web socket connection to the host running at the provided URL.
// `name` here is just an identifying string for the socket, which will be returned when/if the socket is closed
// by calling a provided function (settable with `SetOnClose()`).
// `or` here is the func that's called when a message is read from the socket. The call is made from a separate routine.
// The message types are defined in RFC 6455, section 11.8.
// `oc` here is the func that's called when the socket is just about to be closed. The call is made from a
// separate routine.
// If you do not care about these callbacks, pass nil instead.
func NewClient(name string, socketUrl string, or func(int, []byte), oc func(string)) (S, error) {
	u, err := url.Parse(socketUrl)
	if err != nil {
		log.Println("Could not parse URL from provided URL string:", socketUrl, err)
		return nil, err
	}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		log.Println("Could not connect to provided host:", u.Host, err)
		return nil, err
	}

	ws, _, err := websocket.NewClient(conn, u, nil, 1024, 1024)
	if err != nil {
		log.Println("Error while opening websocket:", err)
		return nil, err
	}

	return socketSetup(name, ws, or, oc), nil
}

func socketSetup(name string, ws *websocket.Conn, or func(int, []byte), oc func(string)) S {
	if or == nil {
		or = func(int, []byte){}
	}

	if oc == nil {
		oc = func(string){}
	}

	s := &s{
		name:      name,
		ws:        ws,
		send:      make(chan []byte, 256),
		shutdown:  make(chan bool),
		onRead:    or,
		onClose:   oc,
	}

	go s.writePump()
	go s.readPump()
	return s
}

func (s *s) SetOnClose(oc func(name string)) {
	if oc != nil {
		s.onClose = oc
	}
}

func (s *s) SetOnRead(or func(messageType int, message []byte)) {
	if or != nil {
		s.onRead = or
	}
}

func (s *s) Write(payload []byte) {
	s.send <- payload
}

func (s *s) Close() {
	s.onClose(s.name)
	s.ws.Close()
}

func (s *s) GetName() string {
	return s.name
}

// readPump pumps messages from the websocket
func (s *s) readPump() {
	for {
		mt, message, err := s.ws.ReadMessage()
		if err != nil {
			// This happens anytime a client closes the connection, which can end up with
			// chatty logs, so we aren't logging this error currently.  If we did, it would look like:
 			// log.Println("[" + s.name + "]", "Error during socket read:", err)
			s.Close()
			s.shutdown <- true
			return
		}

		s.onRead(mt, message)
	}
}

// writePump pumps messages to the websocket
func (s *s) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <- s.shutdown:
			return
		case message := <-s.send:
			if err := s.write(websocket.TextMessage, message); err != nil {
				log.Println("[" + s.name + "]", "Error during socket write:", err)
				s.Close()
				return
			}
		case <-ticker.C:
			if err := s.write(websocket.PingMessage, []byte{}); err != nil {
				log.Println("[" + s.name + "]", "Error during ping for socket:", err)
				s.Close()
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
