// Package geobin.io runs a web server which creates a geobin url that can receive geo data via POSTs and
// visualizes it on a map.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/go-redis/redis"
)

// some read-only global vars
var isDebug = flag.Bool("debug", false, "Boolean flag indicates a debug build. Affects log statements.")
var isVerbose = flag.Bool("verbose", false, "Boolean flag indicates you want to see a lot of log messages.")

func init() {
	// add file info to log statements
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
}

// starts the redis pump and http server
func main() {
	flag.Parse()
	// TODO: verify if this is actually beneficial
	runtime.GOMAXPROCS(runtime.NumCPU())

	// load up config.json
	conf := loadConfig()

	// redis client
	client := redis.NewTCPClient(&redis.Options{
		Addr:     conf.RedisHost,
		Password: conf.RedisPass,
		DB:       conf.RedisDB,
	})

	if ping := client.Ping(); ping.Err() != nil {
		log.Fatal(ping.Err())
	}

	// redis pubsub connection
	ps := client.PubSub()

	// prepare a socketmap
	sm := NewSocketMap(ps)

	// loop for receiving messages from Redis pubsub, and forwarding them on to relevant ws connection
	go redisPump(ps, sm)

	defer func() {
		ps.Close()
		client.Close()
	}()

	// prepare server
	http.Handle("/", NewGeobinServer(conf, client, ps, sm))

	// Start up HTTP server
	log.Println("Starting server at", conf.Host, conf.Port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%d", conf.Host, conf.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// redisPump reads messages out of redis and pushes them through the
// appropriate websocket
func redisPump(ps *redis.PubSub, sm SocketMap) {
	for {
		v, err := ps.Receive()
		if err != nil {
			log.Println("Error from Redis PubSub:", err)
			return
		}

		switch v := v.(type) {
		case *redis.Message:
			if err = sm.Send(v.Channel, []byte(v.Payload)); err != nil {
				log.Println(err)
			}
		}
	}
}

// debugLog logs messages sent to it if and only if isDebug or isVerbose are set to true
func debugLog(v ...interface{}) {
	if *isDebug || *isVerbose {
		log.Println(v...)
	}
}

// verboseLog logs messages sent to it if and only if isVerbose is set to true
func verboseLog(v ...interface{}) {
	if *isVerbose {
		log.Println(v...)
	}
}
