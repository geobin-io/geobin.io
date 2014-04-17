package manager

import (
	"sync"
)

type Manager interface {
	Touch(func(interface{}))
}

type manager struct {
	managed interface{}
	mutex *sync.Mutex
}

func NewManager(managed interface{}) Manager {
	return &manager {
		managed: managed,
		mutex: &sync.Mutex{},
	}
}

func (m *manager) Touch(t func(interface{})) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	t(m.managed)
}
