package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())
}

type geobinServer struct {
	*http.ServeMux
	conf *Config
	RedisClient
	PubSubber
	SocketMap
}

func NewGeobinServer(c *Config, rc RedisClient, ps PubSubber, sm SocketMap) *geobinServer {
	gbs := geobinServer{
		conf:        c,
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
			gb.rateLimit(gb.binHandler, gb.conf.RateLimit)(w, req)
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
	r.HandleFunc("/api/1/create", apiRoute(gb.rateLimit(gb.createHandler, gb.conf.RateLimit)))
	r.HandleFunc("/api/1/history/", apiRoute(gb.rateLimit(gb.historyHandler, gb.conf.RateLimit))) // /api/1/history/{bin_id}
	r.HandleFunc("/api/1/ws/", gb.wsHandler)                                                      // /api/1/ws/{bin_id}

	return r
}

// rateLimit uses redis to enforce rate limits per route. This middleware should
// only be used on routes that contain binIds or other unique identifiers,
// otherwise the rate limit will be globally applied, instead of scoped to a
// particular bin.
func (gb *geobinServer) rateLimit(h http.HandlerFunc, requestsPerSec int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		ts := time.Now().Unix()
		key := fmt.Sprintf("rate-limit:%s:%d", url, ts)

		exists, err := gb.Exists(key)
		if err != nil {
			log.Println(err)
		}

		if exists {
			res, err := gb.RedisClient.Get(key)
			if err != nil {
				http.Error(w, "API Error", http.StatusServiceUnavailable)
				return
			}

			reqCount, _ := strconv.Atoi(res)
			if reqCount >= requestsPerSec {
				http.Error(w, "Rate limit exceeded. Wait a moment and try again.", http.StatusInternalServerError)
				return
			}
		}

		gb.Incr(key)
		gb.Expire(key, 5*time.Second)

		h.ServeHTTP(w, r)
	}
}

// randomString returns a random string with the given length
func (gb *geobinServer) randomString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		b[i] = gb.conf.NameVals[rand.Intn(len(gb.conf.NameVals))]
	}

	s := string(b)

	exists, err := gb.Exists(s)
	if err != nil {
		log.Println("Failure to EXISTS for:", s, err)
		return "", err
	}

	if exists {
		return gb.randomString(length)
	}

	return s, nil
}
