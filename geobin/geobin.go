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

var client = redis.NewTCPClient(&redis.Options{Addr: redisHost,})

type DashBoard struct {
	UUID    string
	History []string
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	r := mux.NewRouter()
	r.HandleFunc("/", create)
	r.HandleFunc("/{uuid}", existing)
	//	r.HandleFunc("/socket/{uuid}", openSocket)
	http.Handle("/", r)
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
	socket := mux.Vars(r)["uuid"]
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

		client.ZAdd(socket, redis.Z{float64(time.Now().UTC().Unix()), string(body)})
		client.Publish(socket, string(body))
		client.Expire(socket, 48*time.Second)
	} else if r.Method == "GET" {
		history := client.ZRevRange(socket, "0", "-1").Val()
		client.Expire(socket, 48*time.Second)
		d := &DashBoard{
			socket,
			history,
		}

		t, _ := template.ParseFiles("html/dashboard.html")
		t.Execute(w, d)
	}
}

//func openSocket(w http.ResponseWriter, r *http.Request) {
//	// upgrade the connection
//	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
//	if _, ok := err.(websocket.HandshakeError); ok {
//		http.Error(w, "Not a websocket handshake", http.StatusBadRequest)
//		return
//	} else if err != nil {
//		log.Println("Error upgrading connection to websocket protocol:", err)
//		http.Error(w, "Error while opening websocket!", http.StatusInternalServerError)
//		return
//	}
//
//	pubsub := client.PubSub()
//	pubsub.Subscribe(socket)
//	pubsub.Receive()
//}

func main() {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
