package socket

import (
	"fmt"
	"github.com/geoloqi/geobin-go/test"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRoundTrip(t *testing.T) {
	var msgCount uint64 = 0
	ts := makeRoundTripServer(t, "test_socket", func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		test.Expect(t, string(message), "You got a message!")
	}, nil)
	roundTrip(t, ts, "test_client")
	test.Expect(t, atomic.LoadUint64(&msgCount), uint64(1))
}

func TestManyRoundTripsManySockets(t *testing.T) {
	var msgCount uint64 = 0
	ts := makeRoundTripServer(t, "test_socket", func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		test.Expect(t, string(message), "You got a message!")
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
	test.Expect(t, atomic.LoadUint64(&msgCount), uint64(count))
}

func TestManyMessagesSingleSocket(t *testing.T) {
	count := 1000
	interval := 100 * time.Microsecond

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sck, err := NewSocket("test_socket", w, r, nil, nil)
		if err != nil {
			t.Error("Error creating websocket:", err)
		}

		go func(s S) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for i := 0; i < count; i++ {
				<-ticker.C
				s.Write([]byte("You got a message!"))
			}
		}(sck)
	}))
	defer ts.Close()

	var msgCount uint64 = 0
	makeClient(t, ts.URL, "test_client", func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		test.Expect(t, string(message), "You got a message!")
	}, nil)

	micros := float64(count) * 1.2
	time.Sleep(time.Duration(micros) * interval)
	test.Expect(t, atomic.LoadUint64(&msgCount), uint64(count))
}

func TestOnClose(t *testing.T) {
	var serverClosed uint64 = 0
	ts := makeRoundTripServer(t, "test_socket", nil, func(name string) {
		atomic.AddUint64(&serverClosed, 1)
		test.Expect(t, name, "test_socket")
	})
	defer ts.Close()

	var clientClosed uint64 = 0
	client := makeClient(t, ts.URL, "test_client", nil, func(name string) {
		atomic.AddUint64(&clientClosed, 1)
		test.Expect(t, name, "test_client")
	})

	client.Close()

	// sleep a lil bit to allow the socket channels to communicate the shut down
	time.Sleep(250 * time.Microsecond)

	test.Expect(t, atomic.LoadUint64(&serverClosed), uint64(1))
	test.Expect(t, atomic.LoadUint64(&clientClosed), uint64(1))
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

func makeClient(t *testing.T, url string, name string, or func(int, []byte), oc func(string)) S {
	client, err := NewClient(name, url, or, oc)
	if err != nil {
		t.Error("Error opening client socket:", name, err)
	}

	return client
}

func roundTrip(t *testing.T, ts *httptest.Server, clientName string) {
	var msgCount uint64 = 0
	client := makeClient(t, ts.URL, clientName, func(messageType int, message []byte) {
		atomic.AddUint64(&msgCount, 1)
		test.Expect(t, string(message), "You got a message!")
	}, nil)
	client.Write([]byte("You got a message!"))

	// sleep a lil bit to allow the server to write back to the websocket
	time.Sleep(25 * time.Millisecond)
	test.Expect(t, atomic.LoadUint64(&msgCount), uint64(1))
}
