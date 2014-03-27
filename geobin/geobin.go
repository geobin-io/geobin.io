package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	gu "github.com/nu7hatch/gouuid"
	redis "github.com/vmihailenco/redis/v2"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	redisHost = "127.0.0.1:6379"
)

var client = redis.NewTCPClient(&redis.Options{Addr: redisHost})
var pubsub = client.PubSub()
var sockets = make(map[string]chan []byte)

type DashBoard struct {
	UUID    string
	History []string
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	r := mux.NewRouter()
	r.HandleFunc("/", create)
	r.HandleFunc("/{uuid}", existing)
	r.HandleFunc("/socket/{uuid}", openSocket)
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
	u, err := gu.NewV4()

	if err != nil {
		log.Println("Failure to generate uuid:", err)
		fmt.Fprint(w, "Could not generate you a new UUID! Please try again.")
		return
	}

	http.Redirect(w, r, "/"+u.String(), http.StatusFound)
}

func existing(w http.ResponseWriter, r *http.Request) {
	uuid := mux.Vars(r)["uuid"]

	if !client.Exists(uuid).Val() {
		http.Error(w, "Unkown UUID, hit index (/) to create one.", http.StatusBadRequest)
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

		var jsonBlob map[string]interface{}
		if err = json.Unmarshal(body, &jsonBlob); err != nil {
			log.Println("Failure to unmarshal POST JSON:", err)
			http.Error(w, "Could not parse body as JSON.", http.StatusBadRequest)
			return
		}

		client.ZAdd(uuid, redis.Z{float64(time.Now().UTC().Unix()), string(body)})
		client.Publish(uuid, string(body))
		client.Expire(uuid, 48*time.Second)
	} else if r.Method == "GET" {
		history := client.ZRevRange(uuid, "0", "-1").Val()
		client.Expire(uuid, 48*time.Second)
		d := &DashBoard{
			uuid,
			history,
		}

		t, _ := template.ParseFiles("templates/dashboard.html")
		t.Execute(w, d)
	}
}
