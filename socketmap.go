package main

import (
	"errors"
	"fmt"
	"sync"
)

type SocketMap interface {
	Add(binName, socketUUID string, s Socket)
	Get(binName, socketUUID string) (Socket, bool)
	Delete(binName, socketUUID string) error
	Send(binName string, payload []byte) error
}

type UnSub interface {
	Unsubscribe(...string) error
}

func NewSocketMap(unsubber UnSub) SocketMap {
	return &sm{
		unsubber: unsubber,
	}
}

type sm struct {
	lk       sync.Mutex
	unsubber UnSub
	smap     map[string]map[string]Socket
}

func (sm *sm) Add(binName, socketUUID string, s Socket) {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.smap == nil {
		sm.smap = make(map[string]map[string]Socket)
	}

	if _, ok := sm.smap[binName]; !ok {
		sm.smap[binName] = make(map[string]Socket)
	}

	sm.smap[binName][socketUUID] = s
}

func (sm *sm) Get(binName, socketUUID string) (Socket, bool) {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.smap == nil {
		return nil, false
	}

	if _, ok := sm.smap[binName]; !ok {
		return nil, false
	}

	return sm.smap[binName][socketUUID], true
}

func (sm *sm) Delete(binName, socketUUID string) error {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.smap == nil {
		return errors.New("There are no known sockets.")
	}

	sockets, ok := sm.smap[binName]
	if !ok {
		return errors.New("No sockets match that bin name.")
	}

	_, ok = sockets[socketUUID]
	if !ok {
		return errors.New("No matching socket for that bin name and uuid.")
	}

	delete(sockets, socketUUID)
	if len(sockets) == 0 {
		delete(sm.smap, binName)

		if sm.unsubber == nil {
			return nil
		}

		if err := sm.unsubber.Unsubscribe(binName); err != nil {
			return errors.New(fmt.Sprint("Failure to UNSUBSCRIBE from", binName, err))
		}
	}
	return nil
}

func (sm *sm) Send(binName string, payload []byte) error {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.smap == nil {
		return errors.New("There are no known sockets.")
	}

	sockets, ok := sm.smap[binName]
	if !ok {
		return errors.New(fmt.Sprint("Got message for unknown channel:", binName))
	}

	for _, s := range sockets {
		go func(s Socket, p []byte) {
			s.Write(p)
		}(s, payload)
	}
	return nil
}
