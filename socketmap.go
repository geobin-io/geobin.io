package main

import (
	"log"
	"sync"

	redis "github.com/vmihailenco/redis/v2"
)

type SocketMap struct {
	lk     sync.Mutex
	PubSub *redis.PubSub
	Map    map[string]map[string]Socket
}

func (sm *SocketMap) Add(binName string, socketUUID string, s Socket) {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.Map == nil {
		sm.Map = make(map[string]map[string]Socket)
	}

	if _, ok := sm.Map[binName]; !ok {
		sm.Map[binName] = make(map[string]Socket)
	}

	sm.Map[binName][socketUUID] = s
}

func (sm *SocketMap) Delete(binName string, socketUUID string) {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.Map == nil {
		return
	}

	sockets, ok := sm.Map[binName]
	if ok {
		delete(sockets, socketUUID)

		if len(sockets) == 0 {
			delete(sm.Map, binName)

			if sm.PubSub == nil {
				return
			}

			if err := sm.PubSub.Unsubscribe(binName); err != nil {
				log.Println("Failure to UNSUBSCRIBE from", binName, err)
			}
		}
	}
}

func (sm *SocketMap) Send(binName string, payload []byte) {
	sm.lk.Lock()
	defer sm.lk.Unlock()
	if sm.Map == nil {
		return
	}

	sockets, ok := sm.Map[binName]
	if !ok {
		log.Println("Got message for unknown channel:", binName)
		return
	}

	for _, s := range sockets {
		go func(s Socket, p []byte) {
			s.Write(p)
		}(s, payload)
	}
}
