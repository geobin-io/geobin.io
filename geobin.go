package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/geoloqi/geobin-go/socket"
	redis "github.com/vmihailenco/redis/v2"
)

var config = &Config{}
var client = &redis.Client{}
var pubsub = &redis.PubSub{}
var socketMap = &SocketMap{
	Map:    make(map[string]map[string]socket.S),
	PubSub: pubsub,
}
var isDebug = flag.Bool("debug", false, "Boolean flag indicates a debug build. Affects log statements.")

// starts the redis pump and http server
func main() {
	flag.Parse()
	// set numprocs
	runtime.GOMAXPROCS(runtime.NumCPU())
	// add file info to log statements
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())

	// prepare router
	r := createRouter()
	http.Handle("/", r)

	loadConfig()
	setupRedis()

	// loop for receiving messages from Redis pubsub, and forwarding them on to relevant ws connection
	go redisPump()

	defer func() {
		pubsub.Close()
		client.Close()
	}()

	// Start up HTTP server
	log.Println("Starting server at", config.Host, config.Port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%d", config.Host, config.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// setupRedis creates a redis client and pubsub
func setupRedis() {
	client = redis.NewTCPClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})

	if ping := client.Ping(); ping.Err() != nil {
		log.Fatal(ping.Err())
	}
	pubsub = client.PubSub()
}

// redisPump reads messages out of redis and pushes them through the
// appropriate websocket
func redisPump() {
	for {
		v, err := pubsub.Receive()
		if err != nil {
			log.Println("Error from Redis PubSub:", err)
			return
		}

		switch v := v.(type) {
		case *redis.Message:
			socketMap.Send(v.Channel, []byte(v.Payload))
		}
	}
}
