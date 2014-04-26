package manager

import "sync"

// Manager provides an nterface for providing thread-safe access to an object
type Manager interface {
	Touch(func(interface{}))
}

type manager struct {
	managed interface{}
	mutex   *sync.Mutex
}

// NewManager creates a new manager that wraps the passed in object
func NewManager(managed interface{}) Manager {
	return &manager{
		managed: managed,
		mutex:   &sync.Mutex{},
	}
}

// Touch allows you to run the given function, with the managed object passed in
func (m *manager) Touch(function func(interface{})) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	function(m.managed)
}
