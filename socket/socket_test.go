package socket

import (
	"github.com/geoloqi/geobin-go/test"
	"testing"
	"net/http/httptest"
	"net/http"
	"time"
	"runtime"
	"sync"
	"fmt"
)

func TestRoundTrip(t *testing.T) {
	ts := makeSocketServer(t)
	roundTrip(t, ts, "test_client")
}

func TestManyRoundTrips(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer runtime.GOMAXPROCS(1)
	ts := makeSocketServer(t)
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
}

func makeSocketServer(t *testing.T) (*httptest.Server){
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sck, err := NewSocket("test_socket", w, r)
		if err != nil {
			t.Error("Error creating websocket:", err)
		}

		sck.Write([]byte("You got a message!"))
	}))

	return ts
}

func makeClient(t *testing.T, url string, name string, or func(int, []byte), oc func(string)) S {
	client, err := NewClient(name, url)
	if err != nil {
		t.Error("Error opening client socket:", name, err)
	}

	client.SetOnRead(or)
	return client
}

func roundTrip(t *testing.T, ts *httptest.Server, clientName string) {
	msgReceived := false
	makeClient(t, ts.URL, clientName, func(messageType int, message []byte) {
		msgReceived = true
		test.Expect(t, string(message), "You got a message!")
	}, nil)

	// sleep a lil bit to allow the server to write back to the websocket
	time.Sleep(25 * time.Millisecond)
	test.Expect(t, msgReceived, true)
}

