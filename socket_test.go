package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

func TestRoundTrip(t *testing.T) {
	var msgCount uint64
	ts := makeRoundTripServer(t, "test_socket", func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		assert.Equal(t, "You got a message!", string(message))
	}, nil)
	roundTrip(t, ts, "test_client")
	assert.Equal(t, uint64(1), atomic.LoadUint64(&msgCount))
}

func TestManyRoundTripsManySockets(t *testing.T) {
	var msgCount uint64
	ts := makeRoundTripServer(t, "test_socket", func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		assert.Equal(t, "You got a message!", string(message))
	}, nil)
	defer ts.Close()

	count := 100
	var w sync.WaitGroup
	w.Add(count)
	for i := 0; i < count; i++ {
		go func(index int) {
			roundTrip(t, ts, fmt.Sprint("test_client:", index))
			w.Done()
		}(i)
	}
	w.Wait()
	assert.Equal(t, uint64(count), atomic.LoadUint64(&msgCount))
}

func TestManyMessagesSingleSocket(t *testing.T) {
	count := 1000
	interval := 100 * time.Microsecond

	var serverMsgReceivedCount uint64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sck, err := NewSocket("test_socket", w, r, func(messageType int, message []byte) {
			atomic.AddUint64(&serverMsgReceivedCount, 1)
			assert.Equal(t, "You got a message!", string(message))
		}, nil)

		if err != nil {
			t.Error("Error creating websocket:", err)
		}

		// when a request comes in, start writing lotsa messages to it
		go writeLotsaMessages(sck, count, interval)
	}))
	defer ts.Close()

	var clientMsgReceivedCount uint64
	client := makeClient(t, ts.URL, "test_client", func(messageType int, message []byte) {
		atomic.AddUint64(&clientMsgReceivedCount, 1)
		assert.Equal(t, "You got a message!", string(message))
	}, nil)

	// we opened a client connection, starting sending lotsa messages to the server
	go writeLotsaMessages(client, count, interval)

	// sleep a bit to let the messages be sent
	time.Sleep(3 * time.Second)

	assert.Equal(t, uint64(count), atomic.LoadUint64(&clientMsgReceivedCount))
	assert.Equal(t, uint64(count), atomic.LoadUint64(&serverMsgReceivedCount))
}

func TestOnClose(t *testing.T) {
	var serverClosed uint64
	ts := makeRoundTripServer(t, "test_socket", nil, func(name string) {
		atomic.AddUint64(&serverClosed, 1)
		assert.Equal(t, "test_socket", name)
	})
	defer ts.Close()

	var clientClosed uint64
	client := makeClient(t, ts.URL, "test_client", nil, func(name string) {
		atomic.AddUint64(&clientClosed, 1)
		assert.Equal(t, "test_client", name)
	})

	client.Close()

	// sleep a lil bit to allow the socket channels to communicate the shut down
	time.Sleep(25 * time.Millisecond)

	assert.Equal(t, uint64(1), atomic.LoadUint64(&serverClosed))
	assert.Equal(t, uint64(1), atomic.LoadUint64(&clientClosed))
}

func makeRoundTripServer(t *testing.T, name string, or func(int, []byte), oc func(string)) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sck, err := NewSocket(name, w, r, or, oc)
		if err != nil {
			t.Error("Error creating websocket:", err)
		}

		sck.Write([]byte("You got a message!"))
	}))

	return ts
}

func makeClient(t *testing.T, url string, name string, or func(int, []byte), oc func(string)) Socket {
	client, err := NewClient(name, url, or, oc)
	if err != nil {
		t.Error("Error opening client socket:", name, err)
	}

	return client
}

func roundTrip(t *testing.T, ts *httptest.Server, clientName string) {
	var msgCount uint64
	client := makeClient(t, ts.URL, clientName, func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		assert.Equal(t, "You got a message!", string(message))
	}, nil)
	client.Write([]byte("You got a message!"))

	// sleep a lil bit to allow the server to write back to the websocket
	time.Sleep(25 * time.Millisecond)
	assert.Equal(t, uint64(1), atomic.LoadUint64(&msgCount))
}

func writeLotsaMessages(sck Socket, count int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for i := 0; i < count; i++ {
		<-ticker.C
		sck.Write([]byte("You got a message!"))
	}
}
