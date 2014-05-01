package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://testing.geobin.io/create", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	createHandler(w, req)

	assertResponseOK(w, t)
	assertBodyContainsKey(w.Body, "id", t)
	assertBodyContainsKey(w.Body, "expires", t)
}

func TestBinHandler404(t *testing.T) {
	// Test 404 for nonexistant bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/nonexistant_bin", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	binHandler(w, req)

	assertResponseNotFound(w, t)
}

func TestBinHandlerEmptyBody(t *testing.T) {
	binId, err := createBin()
	if err != nil {
		t.Error("Could not create bin")
	}

	// Make sure we get an error when we send nothing to our bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	binHandler(w, req)

	assertResponseOK(w, t)
}

func TestBinHandlerNonJSONBody(t *testing.T) {
	binId, err := createBin()
	if err != nil {
		t.Error("Could not create bin")
	}

	// Make sure we get a 200 when we post real data to our bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, strings.NewReader(`deal with it`))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	binHandler(w, req)

	assertResponseOK(w, t)
}

func TestBinHandler200(t *testing.T) {
	binId, err := createBin()
	if err != nil {
		t.Error("Could not create bin")
	}

	// Make sure we get a 200 when we post real data to our bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, strings.NewReader(`{"lat": 10, "lng": -10}`))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	binHandler(w, req)

	assertResponseOK(w, t)
}

/* Test Helpers */

func assertResponseCode(w *httptest.ResponseRecorder, code int, t *testing.T) {
	if w.Code != code {
		t.Errorf("Expected response code %d, got %d", code, w.Code)
	}
}

func assertResponseOK(w *httptest.ResponseRecorder, t *testing.T) {
	assertResponseCode(w, http.StatusOK, t)
}

func assertResponseNotFound(w *httptest.ResponseRecorder, t *testing.T) {
	assertResponseCode(w, http.StatusNotFound, t)
}

func assertBodyContainsKey(body *bytes.Buffer, key string, t *testing.T) {
	var b map[string]interface{}
	json.Unmarshal(body.Bytes(), &b)
	if _, ok := b[key]; !ok {
		t.Error("response doesn't contain '" + key + "'")
	}
}

func createBin() (string, error) {
	req, err := http.NewRequest("GET", "http://testing.geobin.io/create", nil)
	if err != nil {
		return "", err
	}
	w := httptest.NewRecorder()
	createHandler(w, req)

	if w.Code != http.StatusOK {
		return "", errors.New("Error creating bin.")
	}

	var js map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &js)
	id, ok := js["id"]
	if !ok {
		return "", errors.New("Invalid response from /create")
	}

	return id.(string), nil
}
