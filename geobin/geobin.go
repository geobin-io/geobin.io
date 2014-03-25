package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	r := mux.NewRouter()

	http.Handle("/", r)
}

func main() {

}
