package main

import (
	"bytes"
	"github.com/gorilla/mux"
	redis "github.com/vmihailenco/redis/v2"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
	"strings"
)

const (
	redisHost = "127.0.0.1:6379"
	nameVals  = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	nameLen   = 10
)

// todo: determine if these need to be threadsafe
var client = redis.NewTCPClient(&redis.Options{Addr: redisHost})
var pubsub = client.PubSub()
var sockets = make(map[string]chan []byte)

type DashBoard struct {
	Name    string
	History [][]string
}

func init() {
	// add file info to log statements
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())

	// prepare router
	r := mux.NewRouter()
	r.HandleFunc("/create", create)
	r.HandleFunc("/{name}", existing)
	r.HandleFunc("/ws/{name}", openSocket)
	http.Handle("/", r)
}

func main() {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// loop for receiving messages from Redis pubsub, and forwarding them on to relevant ws connection
	for {
		v, err := pubsub.Receive()
		if err != nil {
			log.Println("Error from Redis PubSub:", err)
			return
		}

		switch v := v.(type) {
		case redis.Message:
			wsChan, ok := sockets[v.Channel]
			if !ok {
				log.Println("Got message for unknown channel:", v.Channel)
				return
			}

			wsChan <- []byte(v.Payload)
		}
	}

	pubsub.Close()
	client.Close()
}

func create(w http.ResponseWriter, r *http.Request) {
	n, err := randomString(nameLen)
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

	if res := client.Expire(n, 48*time.Hour); res.Err() != nil {
		log.Println("Failure to set EXPIRE for", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+n, http.StatusFound)
}

func existing(w http.ResponseWriter, r *http.Request) {
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

	if r.Method == "POST" {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error while reading POST body:", err)
			http.Error(w, "Could not read POST body!", http.StatusInternalServerError)
			return
		}

		// put headers into the set member to make the value unique
		var buffer bytes.Buffer
		for k, v := range r.Header {
			buffer.WriteString(k + ": " + strings.Join(v, ", ") + "\r\n")
		}
		buffer.WriteString(string(body))

		if res := client.ZAdd(name, redis.Z{float64(time.Now().UTC().Unix()), buffer.String()}); res.Err() != nil {
			log.Println("Failure to ZADD to", name, res.Err())
		}

		if res := client.Publish(name, buffer.String()); res.Err() != nil {
			log.Println("Failure to PUBLISH to", name, res.Err())
		}
	} else if r.Method == "GET" {
		set := client.ZRevRange(name, "0", "-1")
		if set.Err() != nil {
			log.Println("Failure to ZREVRANGE for", name, set.Err())
		}

		// chop off the last history member since it is the placeholder value from when the set was created
		vals := set.Val()[:len(set.Val()) - 1]

		history := make([][]string, len(vals))
		for i, v := range vals {
			history[i] = strings.Split(v, "\r\n")
		}

		// dashboard context object,
		d := &DashBoard{
			name,
			history,
		}

		t, _ := template.ParseFiles("templates/dashboard.html")
		t.Execute(w, d)
	}
}

func randomString(length int) (string, error) {
	b := make([]byte, length)
	for i, _ := range b {
		b[i] = nameVals[rand.Intn(len(nameVals))]
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
	} else {
		if resp.Val() {
			return true, nil
		} else {
			return false, nil
		}
	}
}
