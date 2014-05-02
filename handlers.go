package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	redis "github.com/vmihailenco/redis/v2"
)

func createRouter() *http.ServeMux {
	r := http.NewServeMux()

	// Web routes
	r.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case "GET":
			debugLog("web -", req.URL)
			http.ServeFile(w, req, "static/app/index.html")
		case "POST":
			binHandler(w, req)
		}
	})
	r.HandleFunc("/static/", func(w http.ResponseWriter, req *http.Request) {
		debugLog("static -", req.URL)
		http.ServeFile(w, req, req.URL.Path[1:])
	})

	// API routes
	r.HandleFunc("/api/1/create", createHandler)
	r.HandleFunc("/api/1/history/", historyHandler) // /api/1/history/{bin_id}
	r.HandleFunc("/api/1/ws/", wsHandler)           // /api/1/ws/{bin_id}

	return r
}

// Creates a new bin
func createHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("create -", r.URL)

	// Get a new name
	n, err := randomString(config.NameLength)
	if err != nil {
		log.Println("Failure to create new name:", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	// Save to redis
	if res := client.ZAdd(n, redis.Z{Score: 0, Member: ""}); res.Err() != nil {
		log.Println("Failure to ZADD to", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	// Set expiration
	d := 48 * time.Hour
	if res := client.Expire(n, d); res.Err() != nil {
		log.Println("Failure to set EXPIRE for", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}
	exp := time.Now().Add(d).Unix()

	// Create the json response and encoder
	encoder := json.NewEncoder(w)
	bin := map[string]interface{}{
		"id":      n,
		"expires": exp,
	}

	// encode the json directly to the response writer
	err = encoder.Encode(bin)
	if err != nil {
		log.Println("Failure to create json for new name:", n, err)
		http.Error(w, fmt.Sprintf("New Geobin created (%v) but we could not return the JSON for it!", n), http.StatusInternalServerError)
		return
	}
}

// log a request into a bin
func binHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("bin -", r.URL)
	name := r.URL.Path[1:]

	exists, err := nameExists(name)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.NotFound(w, r)
		return
	}

	var body []byte
	if r.Body != nil {
		body, err = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("Error while reading POST body:", err)
			http.Error(w, "Could not read POST body!", http.StatusInternalServerError)
			return
		}
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	gr := NewGeobinRequest(time.Now().UTC().Unix(), headers, body)

	encoded, err := json.Marshal(gr)
	if err != nil {
		log.Println("Error marshalling request:", err)
	}

	if res := client.ZAdd(name, redis.Z{Score: float64(time.Now().UTC().Unix()), Member: string(encoded)}); res.Err() != nil {
		log.Println("Failure to ZADD to", name, res.Err())
	}

	if res := client.Publish(name, string(encoded)); res.Err() != nil {
		log.Println("Failure to PUBLISH to", name, res.Err())
	}
}

// Get bin history
func historyHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("history -", r.URL)
	path := strings.Split(r.URL.Path, "/")
	name := path[len(path)-1]

	exists, err := nameExists(name)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.NotFound(w, r)
		return
	}

	set := client.ZRevRange(name, "0", "-1")
	if set.Err() != nil {
		log.Println("Failure to ZREVRANGE for", name, set.Err())
	}

	// chop off the last history member since it is the placeholder value from when the set was created
	vals := set.Val()[:len(set.Val())-1]

	history := make([]GeobinRequest, 0, len(vals))
	for _, v := range vals {
		var gr GeobinRequest
		if err := json.Unmarshal([]byte(v), &gr); err != nil {
			log.Println("Error unmarshalling request history:", err)
		}
		history = append(history, gr)
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(history)
	if err != nil {
		log.Println("Error marshalling request history:", err)
		http.Error(w, "Could not generate history.", http.StatusInternalServerError)
		return
	}
}

// Web socket connections
func wsHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("create -", r.URL)
	path := strings.Split(r.URL.Path, "/")
	binName := path[len(path)-1]

	// start pub subbing
	if err := pubsub.Subscribe(binName); err != nil {
		log.Println("Failure to SUBSCRIBE to", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := uuid.NewV4()
	if err != nil {
		log.Println("Failure to generate new socket UUID", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	uuid := id.String()

	s, err := NewSocket(binName+"~br~"+uuid, w, r, nil, func(socketName string) {
		// the socketname is a composite of the bin name, and the socket UUID
		ids := strings.Split(socketName, "~br~")
		bn := ids[0]
		suuid := ids[1]
		if err := socketMap.Delete(bn, suuid); err != nil {
			log.Println(err)
		}
	})

	if err != nil {
		// if there is an error, NewSocket will have already written a response via http.Error()
		// so only write a log
		log.Println("Error opening websocket:", err)
		return
	}

	// keep track of the outbound channel for pubsubbery
	socketMap.Add(binName, uuid, s)
}
