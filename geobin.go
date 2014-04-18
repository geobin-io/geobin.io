package main

import (
	"encoding/json"
	"fmt"
	"github.com/geoloqi/geobin-go/socket"
	"github.com/gorilla/mux"
	gu "github.com/nu7hatch/gouuid"
	redis "github.com/vmihailenco/redis/v2"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port       int
	RedisHost  string
	RedisPass  string
	RedisDB    int64
	NameVals   string
	NameLength int
}

// TODO: determine if these need to be threadsafe (pretty sure they do)
var config = &Config{}
var client = &redis.Client{}
var pubsub = &redis.PubSub{}
var sockets = make(map[string]map[string]socket.S)

type GeobinRequest struct {
	Timestamp int64             `json:"timestamp"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

func main() {
	// loop for receiving messages from Redis pubsub, and forwarding them on to relevant ws connection
	go redisPump()

	defer func() {
		pubsub.Close()
		client.Close()
	}()

	// Start up HTTP server
	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/*
 * Initilization
 */
func init() {
	// add file info to log statements
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())

	// prepare router
	r := createRouter()
	http.Handle("/", r)

	loadConfig()
	setupRedis()
}

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}

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

func createRouter() *mux.Router {
	r := mux.NewRouter()
	// API routes (POSTs only!)
	api := r.Methods("POST").PathPrefix("/api/{v:[0-9.]+}/").Subrouter()
	api.HandleFunc("/create", createHandler)
	api.HandleFunc("/history/{name}", historyHandler)
	api.HandleFunc("/ws/{name}", wsHandler)

	// Client/web requests (GETs only!)
	web := r.Methods("GET").Subrouter()
	// Any GET request to the /api/ route will serve up the docs static site directly.
	web.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "docs - %v\n", req.URL)
		// TODO: This is wrong, will fix when we actually have the files to serve
		http.ServeFile(w, req, "docs/build/")
	})
	// Any GET request to the /static/ route will serve the files in the static dir directly.
	web.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "static - %v\n", req.URL)
		http.ServeFile(w, req, req.URL.Path[1:])
	})	
	// All other GET requests will serve up the Angular app at static/index.html
	web.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "web - %v\n", req.URL)
		http.ServeFile(w, req, "static/index.html")
	})
	return r
}

/*
 * API Routes
 */
func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "create - %v\n", r.URL)
	n, err := randomString(config.NameLength)
	if err != nil {
		log.Println("Failure to create new name:", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	if res := client.ZAdd(n, redis.Z{0, ""}); res.Err() != nil {
		log.Println("Failure to ZADD to", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	d := 48*time.Hour
	if res := client.Expire(n, d); res.Err() != nil {
		log.Println("Failure to set EXPIRE for", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}
	exp := time.Now().Add(d).Unix()

	bin := map[string]interface{} {
		"id": n,
		"expires": exp,
	}
	binJson, err := json.Marshal(bin)
	if err != nil {
		log.Println("Failure to create json for new name:", n, err)
		// I know this error message is ridiculous, but I don't know how this would ever happen so...
		http.Error(w, fmt.Sprintf("New Geobin created (%v) but we could not make a JSON object for it!", n), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(binJson); err != nil {
		http.Error(w, fmt.Sprintf("New Geobin created (%v) but we failed to write to the response!", n), http.StatusInternalServerError)
		return
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "history - %v", r.URL)
	name := mux.Vars(r)["name"]
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

	resp, err := json.Marshal(history)
	if err != nil {
		log.Println("Error marshalling request history:", err)
		http.Error(w, "Could not generate history.", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(resp))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "create - %v", r.URL)
	// upgrade the connection
	binName := mux.Vars(r)["name"]

	// start pub subbing
	if err := pubsub.Subscribe(binName); err != nil {
		log.Println("Failure to SUBSCRIBE to", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := gu.NewV4()
	if err != nil {
		log.Println("Failure to generate new socket UUID", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	uuid := id.String()

	s, err := socket.NewSocket(binName+"~br~"+uuid, w, r)
	if err != nil {
		// if there is an error, NewSocket will have already written a response via http.Error()
		// so only write a log
		log.Println("Error opening websocket:", err)
		return
	}

	s.SetOnClose(func(socketName string) {
		// the socketname is a composite of the bin name, and the socket UUID
		ids := strings.Split(socketName, "~br~")
		bn := ids[0]
		suuid := ids[1]

		socks, ok := sockets[bn]
		if ok {
			delete(socks, suuid)

			if len(socks) == 0 {
				delete(sockets, bn)
				if err := pubsub.Unsubscribe(bn); err != nil {
					log.Println("Failure to UNSUBSCRIBE from", bn, err)
				}
			}
		}
	})

	// keep track of the outbound channel for pubsubbery
	if _, ok := sockets[binName]; !ok {
		sockets[binName] = make(map[string]socket.S)
	}
	sockets[binName][uuid] = s
}

/*
 * Redis
 */
func redisPump() {
	for {
		v, err := pubsub.Receive()
		if err != nil {
			log.Println("Error from Redis PubSub:", err)
			return
		}

		switch v := v.(type) {
		case *redis.Message:
			sockMap, ok := sockets[v.Channel]
			if !ok {
				log.Println("Got message for unknown channel:", v.Channel)
				return
			}

			for _, sock := range sockMap {
				go func(s socket.S, p []byte) {
					s.Write(p)
				}(sock, []byte(v.Payload))
			}
		}
	}
}

/*
 * Utils
 */
func randomString(length int) (string, error) {
	b := make([]byte, length)
	for i, _ := range b {
		b[i] = config.NameVals[rand.Intn(len(config.NameVals))]
	}

	s := string(b)

	exists, err := nameExists(s)
	if err != nil {
		log.Println("Failure to EXISTS for:", s, err)
		return "", err
	}

	if exists {
		return randomString(length)
	}

	return s, nil
}

func nameExists(name string) (bool, error) {
	resp := client.Exists(name)
	if resp.Err() != nil {
		return false, resp.Err()
	}

	return resp.Val(), nil
}
