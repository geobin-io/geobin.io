package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmizerany/assert"
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

	// Make sure we accept null post bodies
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

func TestBinHistoryReturnsErrorForInvalidBin(t *testing.T) {
	binId := "neverland"

	// Check history for our bin
	req, err := http.NewRequest("GET", "http://testing.geobin.io/api/1/history/"+binId, nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	historyHandler(w, req)

	assertResponseCode(w, http.StatusNotFound, t)
}

func TestBinHistoryWorksAsIntended(t *testing.T) {
	binId, err := createBin()
	if err != nil {
		t.Error("Could not create bin")
	}

	// Post some stuff to the bin
	payload := `{"lat": 10, "lng": -10}`
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, strings.NewReader(payload))
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	binHandler(w, req)

	// Check history for our bin
	req, err = http.NewRequest("GET", "http://testing.geobin.io/api/1/history/"+binId, nil)
	if err != nil {
		t.Error(err)
	}

	w = httptest.NewRecorder()
	historyHandler(w, req)

	assertResponseOK(w, t)

	var history []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
		t.Error(err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 result, got %d", len(history))
	}

	geo := history[0]

	if payload != geo["body"].(string) {
		t.Errorf("Expected: %s\nGot: %s", payload, geo["body"].(string))
	}
}

func TestCountsWorksAsIntended(t *testing.T) {
	bins, expected := createBins([]int{1, 0, 5, 23}, t)

	verifyCounts(bins, expected, t)
}

func TestCountsWithInvalidBinId(t *testing.T) {
	bins, expected := createBins([]int{1, 0, 2}, t)

	bins = append(bins, "invalid")
	expected["invalid"] = nil

	verifyCounts(bins, expected, t)
}

func createBins(counts []int, t *testing.T) (binIds []string, expected map[string]interface{}) {
	binIds = make([]string, len(counts))
	expected = make(map[string]interface{})

	// create the bins
	for i, c := range counts {
		binId, err := createBin()
		if err != nil {
			t.Error(err)
		}

		binIds[i] = binId
		expected[binId] = float64(c)

		// send the appropriate number of requests
		for reqNum := 0; reqNum < c; reqNum++ {
			req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, strings.NewReader(fmt.Sprintf(`{"lat": 10, "lng": -10, "reqNum": %d}`, reqNum)))
			if err != nil {
				t.Error(err)
			}
			req.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			binHandler(w, req)
			assertResponseOK(w, t)
		}
	}
	return
}

func verifyCounts(bins []string, expected map[string]interface{}, t *testing.T) {
	// make the request to the counts route
	w := httptest.NewRecorder()
	var binJson []byte
	var err error
	if binJson, err = json.Marshal(bins); err != nil {
		t.Error(err)
	}
	req, err := http.NewRequest("POST", "http://testing.geobin.io/api/1/counts", strings.NewReader(string(binJson)))
	if err != nil {
		t.Error(err)
	}
	countsHandler(w, req)
	assertResponseOK(w, t)

	// verify the counts are correct
	var got map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &got)
	assert.Equal(t, expected, got)
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
