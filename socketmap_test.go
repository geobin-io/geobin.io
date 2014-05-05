package main

import (
	"sync"
	"testing"
	"time"

	"github.com/bmizerany/assert"
)

type unsubFunc func(...string) error

func (us unsubFunc) Unsubscribe(channels ...string) error {
	return us(channels...)
}

type MockSocket struct {
	lk       sync.Mutex
	name     string
	didWrite bool
}

func (ms *MockSocket) Write(payload []byte) {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	ms.didWrite = true
}

func (ms *MockSocket) getDidWrite() bool {
	ms.lk.Lock()
	defer ms.lk.Unlock()
	return ms.didWrite
}

func (ms *MockSocket) GetName() string {
	return ms.name
}

func (ms *MockSocket) Close() {}

func TestNewSocketMap(t *testing.T) {
	sm := NewSocketMap(getUnsubFunc(t))
	assert.NotEqual(t, nil, sm)
}

func TestAddAndGet(t *testing.T) {
	sm := NewSocketMap(getUnsubFunc(t))
	ms := &MockSocket{name: "mock_socket"}
	sm.Add("bin_name", "socket_uuid", ms)
	sck, ok := sm.Get("bin_name", "socket_uuid")
	assert.Equal(t, true, ok)
	assert.Equal(t, "mock_socket", sck.GetName())
}

func TestDelete(t *testing.T) {
	var didUnsub bool = false
	unsubf := func(channels ...string) error {
		assert.Equal(t, "bin_name", channels[0])
		didUnsub = true
		return nil
	}

	sm := NewSocketMap(unsubFunc(unsubf))
	err := sm.Delete("bin_name", "socket_uuid1")
	assert.NotEqual(t, nil, err)
	ms1 := &MockSocket{name: "mock_socket1"}
	ms2 := &MockSocket{name: "mock_socket2"}
	sm.Add("bin_name", "socket_uuid1", ms1)
	sm.Add("bin_name", "socket_uuid2", ms2)

	err = sm.Delete("bin_name", "unknown_uuid")
	assert.NotEqual(t, nil, err)
	err = sm.Delete("unknown_bin_name", "socket_uuid1")
	assert.NotEqual(t, nil, err)
	err = sm.Delete("bin_name", "socket_uuid1")
	assert.Equal(t, nil, err)
	assert.Equal(t, false, didUnsub)
	err = sm.Delete("bin_name", "socket_uuid2")
	assert.Equal(t, nil, err)
	assert.Equal(t, true, didUnsub)
}

func TestSend(t *testing.T) {
	sm := NewSocketMap(getUnsubFunc(t))
	err := sm.Send("bin_name", []byte("a message"))
	assert.NotEqual(t, nil, err)
	ms := &MockSocket{name: "mock_socket"}

	sm.Add("bin_name", "socket_uuid", ms)
	err = sm.Send("unknown_bin_name", []byte("a message"))
	assert.NotEqual(t, nil, err)
	err = sm.Send("bin_name", []byte("a message"))
	assert.Equal(t, nil, err)
	// sleep a bit to allow go routines to be scheduled and run
	time.Sleep(25 * time.Microsecond)
	assert.Equal(t, true, ms.getDidWrite())
}

func getUnsubFunc(t *testing.T) unsubFunc {
	us := func(channels ...string) error {
		t.Error("Unexpected call to unsubscibe!")
		return nil
	}

	return unsubFunc(us)
}
