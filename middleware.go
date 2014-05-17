package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// rateLimit uses redis to enforce rate limits per route. This middleware should
// only be used on routes that contain binIds or other unique identifiers,
// otherwise the rate limit will be globally applied, instead of scoped to a
// particular bin.
func rateLimit(h http.HandlerFunc, requestsPerSec int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Path
		ts := time.Now().Unix()
		key := fmt.Sprintf("rate-limit:%s:%d", url, ts)

		exists, err := nameExists(key)
		if err != nil {
			log.Println(err)
		}
		if exists {
			res, err := client.Get(key).Result()
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

		tx := client.Multi()
		tx.Incr(key)
		tx.Expire(key, 5*time.Second)
		tx.Close()

		h.ServeHTTP(w, r)
	}
}
