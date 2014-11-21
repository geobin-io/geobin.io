package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/nu7hatch/gouuid"
)

func (gb *geobinServer) createBin(n string, w http.ResponseWriter) (time.Time, error) {
	var err error
	t := time.Now()

	// Save to redis
	if _, err = gb.ZAdd(n, redis.Z{Score: 0, Member: ""}); err != nil {
		log.Println("Failure to ZADD to", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return t, err
	}

	// Set expiration
	d := 48 * time.Hour
	if _, err = gb.Expire(n, d); err != nil {
		log.Println("Failure to set EXPIRE for", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return t, err
	}

	return t.Add(d), nil
}

// createHandler handles requests to /api/1/create. It creates a randomly generated bin_id,
// creates an entry in redis for it, with a 48 hour expiration time and writes a json object
// to the response with the following structure:
//
// `{
//    "id": {bin_id},
//    "expires": {expiration_timestamp}
// }`
//
// The expiration timestamp is in Unix time (milis).
func (gb *geobinServer) createHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("create -", r.URL)

	// Get a new name
	n, err := gb.randomString(gb.conf.NameLength)
	if err != nil {
		log.Println("Failure to create new name:", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	var t time.Time
	if t, err = gb.createBin(n, w); err != nil {
		return
	}

	// Create the json response and encoder
	encoder := json.NewEncoder(w)
	bin := map[string]interface{}{
		"id":      n,
		"expires": t.Unix(),
	}

	// encode the json directly to the response writer
	err = encoder.Encode(bin)
	if err != nil {
		log.Println("Failure to create json for new name:", n, err)
		http.Error(w, fmt.Sprintf("New Geobin created (%v) but we could not return the JSON for it!", n), http.StatusInternalServerError)
		return
	}
}

// countsHandler handles requests to /api/1/counts. It requires an array of binIds as input
// and responds with a dictionary with the binIds as the key and the number of requests stored
// in the db for that binId. If a binId is not found in the db, the value for that binId in the
// response will be null.
func (gb *geobinServer) countsHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("counts -", r.URL)

	// get list of binIds from request body
	var binIds []string
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&binIds); err != nil {
		log.Println("Error marshalling request:", err)
		http.Error(w, "Error marshalling request:", http.StatusBadRequest)
	}

	// look up each binId in db
	counts := make(map[string]interface{})
	for _, binId := range binIds {
		if c, err := gb.ZCount(binId, "-inf", "+inf"); err == nil && c > 0 {
			counts[binId] = c - 1
		} else {
			counts[binId] = nil
		}
	}

	// return counts
	if err := json.NewEncoder(w).Encode(counts); err != nil {
		log.Println("Error encoding response:", err)
		http.Error(w, "Error encoding response!", http.StatusInternalServerError)
	}
}

// binHandler handles requests to /api/1/{binId}. It requires a binId in the request path and some
// JSON in the POST body. It creates a new GeobinRequest object using the body, which in turn
// searches for any geo data in said JSON. It then adds the hydrated GeobinRequest to the database.
func (gb *geobinServer) binHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("bin -", r.URL)
	name := r.URL.Path[1:]

	exists, err := gb.Exists(name)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !exists {
		// TODO: Create bin.
		http.NotFound(w, r)
		return
	}

	var body []byte
	if r.Body != nil {
		// Limit reading of the request body to the first 1MB (1<<20 bytes)
		body, err = ioutil.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		defer r.Body.Close()
		if err != nil {
			log.Println("Error while reading POST body:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if _, err = gb.ZAdd(name, redis.Z{Score: float64(time.Now().UTC().Unix()), Member: string(encoded)}); err != nil {
		log.Println("Failure to ZADD to", name, err)
	}

	if _, err = gb.Publish(name, string(encoded)); err != nil {
		log.Println("Failure to PUBLISH to", name, err)
	}
}

// historyHandler handles requests to /api/v1/history/{bin_id}. It requires a bin_id in the
// request path. It looks said bin_id up in the database and writes all of the GeobinRequests in
// the database for that bin_id to the response as JSON.
func (gb *geobinServer) historyHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("history -", r.URL)
	path := strings.Split(r.URL.Path, "/")
	name := path[len(path)-1]

	exists, err := gb.Exists(name)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if !exists {
		// TODO: Create new bin with given binId
		http.NotFound(w, r)
		return
	}

	set, err := gb.ZRevRange(name, "0", "-1")
	if err != nil {
		log.Println("Failure to ZREVRANGE for", name, err)
	}

	// chop off the last history member since it is the placeholder value from when the set was created
	vals := set[:len(set)-1]

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

// wsHandler handles requests to /api/1/ws/{bin_id}. It requires a bin_id in the request path
// and it subscribes to listen for changes to the bin_id in redis. It creates a socket with
// a UUID and adds that socket to the socketMap. It then sends any updates to the bin_id in
// redis to the socket as they come in.
func (gb *geobinServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	debugLog("create -", r.URL)
	path := strings.Split(r.URL.Path, "/")
	binName := path[len(path)-1]

	// TODO: Check if binId exists, create it if it doesn't.

	// start pub subbing
	if err := gb.Subscribe(binName); err != nil {
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
		if err := gb.Delete(bn, suuid); err != nil {
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
	gb.Add(binName, uuid, s)
}
