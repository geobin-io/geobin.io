package main

import (
	"fmt"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	redis "github.com/vmihailenco/redis/v2"
	"log"
	"net/http"
)

const (
	redisHost = "127.0.0.1"
	redisPass = ""
	redisDB   = 0
)

var client = redis.NewTCPClient(&redis.Options{Addr: redisHost, Password: redisPass, DB: redisDB})

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	r := mux.NewRouter()
	r.HandleFunc("/", newSocket)
	r.HandleFunc("/{socket}", existingSocket)
	http.Handle("/", r)
}

func newSocket(w http.ResponseWriter, r *http.Request) {
	u, err := uuid.NewV4()

	if err != nil {
		log.Println("Failure to generate uuid:", err)
		fmt.Fprintf(w, "Could not generate you a new UUID! Please try again.")
	}

	http.Redirect(w, r, "/"+u.String(), http.StatusFound)
}

func existingSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if r.Method == "POST" {
		// this is a callback
		fmt.Fprintln(w, "You are a callback!", vars["socket"])
	} else if r.Method == "GET" {
		fmt.Fprintln(w, "You are a browser!", vars["socket"])
	}
}

func main() {
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
