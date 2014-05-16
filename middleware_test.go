package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}

	limited := rateLimit(handler, 1)

	req, _ := http.NewRequest("GET", "http://testing.geobin.io/", nil)
	wOk := httptest.NewRecorder()
	wErr := httptest.NewRecorder()

	// Rate limit is 1/s, first one should get a 200
	limited(wOk, req)
	assertResponseOK(wOk, t)

	// Second request should get a 500
	limited(wErr, req)
	assertResponseCode(wErr, 500, t)
}
