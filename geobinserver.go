package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

// requests per second
const limit = 1

func init() {
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())
}

// mock our use of redis pubsub for modularity/testing purposes
type PubSubber interface {
	Subscribe(channels ...string) error
	Unsubscribe(channels ...string) error
}

// mock our use of redis client for modularity/testing purposes
type RedisClient interface {
	ZAdd(key string, members ...redis.Z) *redis.IntCmd
	ZCount(key, min, max string) *redis.IntCmd
	Expire(key string, dur time.Duration) *redis.BoolCmd
	Publish(channel, message string) *redis.IntCmd
	ZRevRange(key, start, stop string) *redis.StringSliceCmd
	Exists(key string) *redis.BoolCmd
	Get(key string) *redis.StringCmd
	Multi() *redis.Multi
}

type geobinServer struct {
	*http.ServeMux
	RedisClient
	PubSubber
	SocketMap
}

func NewGeobinServer(rc RedisClient, ps PubSubber, sm SocketMap) *geobinServer {
	gbs := geobinServer{
		RedisClient: rc,
		PubSubber:   ps,
		SocketMap:   sm,
	}

	gbs.ServeMux = gbs.createRouter()
	return &gbs
}

// createRouter creates the http.HandleFunc to route requests to the handlers defined below.
func (gb *geobinServer) createRouter() *http.ServeMux {
	r := http.NewServeMux()

	// Web routes
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			debugLog("web -", req.URL)
			http.ServeFile(w, req, "static/app/index.html")
		case "POST":
			gb.rateLimit(gb.binHandler, limit)(w, req)
		}
	})
	r.HandleFunc("/static/", func(w http.ResponseWriter, req *http.Request) {
		debugLog("static -", req.URL)
		http.ServeFile(w, req, req.URL.Path[1:])
	})

	// API routes
	// this wraps all of the API routes and checks to see if the request is a POST, if it's not, it forwards the request to the angular app
	apiRoute := func(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, req *http.Request) {
			switch req.Method {
			case "POST":
				h(w, req)
			case "GET":
				debugLog("api (GET) -", req.URL)
				http.ServeFile(w, req, "static/app/index.html")
			}
		}
	}

	r.HandleFunc("/api/1/counts", apiRoute(gb.countsHandler))
	r.HandleFunc("/api/1/create", apiRoute(gb.rateLimit(gb.createHandler, limit)))
	r.HandleFunc("/api/1/history/", apiRoute(gb.rateLimit(gb.historyHandler, limit))) // /api/1/history/{bin_id}
	r.HandleFunc("/api/1/ws/", gb.wsHandler)                                          // /api/1/ws/{bin_id}

	return r
}

// randomString returns a random string with the given length
func (gb *geobinServer) randomString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		b[i] = config.NameVals[rand.Intn(len(config.NameVals))]
	}

	s := string(b)

	exists, err := gb.nameExists(s)
	if err != nil {
		log.Println("Failure to EXISTS for:", s, err)
		return "", err
	}

	if exists {
		return gb.randomString(length)
	}

	return s, nil
}

// nameExists returns true if the specified bin name exists
func (gb *geobinServer) nameExists(name string) (bool, error) {
	resp := gb.Exists(name)
	if resp.Err() != nil {
		return false, resp.Err()
	}

	return resp.Val(), nil
}
