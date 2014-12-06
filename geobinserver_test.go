package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bmizerany/assert"
	"github.com/go-redis/redis"
)

var testConf = &Config{
	"localhost",
	8080,
	"127.0.0.1:6379",
	"",
	0,
	"023456789abcdefghjkmnopqrstuvwxyzABCDEFGHJKMNOPQRSTUVWXYZ",
	10,
	999,
}

type MockRedis struct {
	sync.Mutex
	bins  map[string][]string
	incrs map[string]string
}

type MockPubSub struct{}

func NewMockRedis() *MockRedis {
	return &MockRedis{
		bins:  make(map[string][]string),
		incrs: make(map[string]string),
	}
}

func (mr *MockRedis) ZAdd(key string, members ...redis.Z) (int64, error) {
	mr.Lock()
	defer mr.Unlock()

	if _, ok := mr.bins[key]; !ok {
		mr.bins[key] = make([]string, 0, 0)
	}

	for _, v := range members {
		mr.bins[key] = append(mr.bins[key], v.Member)
	}
	return int64(len(mr.bins[key])), nil
}

func (mr *MockRedis) ZCount(key, min, max string) (int64, error) {
	mr.Lock()
	defer mr.Unlock()
	if _, ok := mr.bins[key]; !ok {
		return 0, errors.New("No bin by that id.")
	}
	return int64(len(mr.bins[key])), nil
}

func (mr *MockRedis) Expire(key string, dur time.Duration) (bool, error) {
	return true, nil
}

func (mr *MockRedis) Publish(channel, message string) (int64, error) {
	return 1, nil
}

func (mr *MockRedis) ZRevRange(key, start, stop string) ([]string, error) {
	mr.Lock()
	defer mr.Unlock()
	if _, ok := mr.bins[key]; !ok {
		return nil, errors.New("No bin by that id.")
	}

	cur := mr.bins[key]
	reversed := make([]string, len(cur), cap(cur))
	copy(reversed, cur)
	sort.Sort(sort.Reverse(sort.StringSlice(reversed)))

	return reversed, nil
}

func (mr *MockRedis) Exists(key string) (bool, error) {
	mr.Lock()
	defer mr.Unlock()
	_, bin := mr.bins[key]
	_, incr := mr.incrs[key]

	return bin || incr, nil
}

// currently Get() is only used for checking on integer values during ratelimiting
// so we don't have to worry about checking mr.bins here, or converting between
// []string and string
func (mr *MockRedis) Get(key string) (string, error) {
	mr.Lock()
	defer mr.Unlock()

	if incr, iok := mr.incrs[key]; iok {
		return incr, nil
	}

	return "", errors.New("No value for key!")
}

func (mr *MockRedis) Incr(key string) (int64, error) {
	mr.Lock()
	defer mr.Unlock()
	if _, ok := mr.incrs[key]; !ok {
		mr.incrs[key] = "1"
		return 1, nil
	}

	incrd, _ := strconv.Atoi(mr.incrs[key])
	incrd++
	mr.incrs[key] = strconv.Itoa(incrd)
	return int64(incrd), nil
}

func (mps *MockPubSub) Subscribe(channels ...string) error {
	return nil
}

func (mps *MockPubSub) Unsubscribe(channels ...string) error {
	return nil
}

func TestCreateHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "http://testing.geobin.io/api/1/create", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	createGeobinServer().ServeHTTP(w, req)

	assertResponseOK(w, t)
	assertBodyContainsKey(w.Body, "id", t)
	assertBodyContainsKey(w.Body, "expires", t)
}

func TestCustomBinHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "http://testing.geobin.io/my_awesome_bin_id", nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	createGeobinServer().ServeHTTP(w, req)

	assertResponseOK(w, t)
}

func TestBinHandlerEmptyBody(t *testing.T) {
	gbs := createGeobinServer()
	binId, err := createBin(gbs)
	if err != nil {
		t.Error("Could not create bin")
	}

	// Make sure we accept null post bodies
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	gbs.ServeHTTP(w, req)

	assertResponseOK(w, t)
}

func TestBinHandlerNonJSONBody(t *testing.T) {
	gbs := createGeobinServer()
	binId, err := createBin(gbs)
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
	gbs.ServeHTTP(w, req)

	assertResponseOK(w, t)
}

func TestBinHandler200(t *testing.T) {
	gbs := createGeobinServer()
	binId, err := createBin(gbs)
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
	gbs.ServeHTTP(w, req)

	assertResponseOK(w, t)
}

func TestBinHistoryCreatesNonExistantBin(t *testing.T) {
	binId := "neverland"

	// Check history for our bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/api/1/history/"+binId, nil)
	if err != nil {
		t.Error(err)
	}

	w := httptest.NewRecorder()
	createGeobinServer().ServeHTTP(w, req)

	assertResponseOK(w, t)
}

func TestBinHistoryWorksAsIntended(t *testing.T) {
	gbs := createGeobinServer()
	binId, err := createBin(gbs)
	if err != nil {
		t.Error("Could not create bin")
	}

	// Post some stuff to the bin
	payload := `{"lat": 10, "lng": -10}`

	_, err = postToBin(gbs, binId, payload)
	if err != nil {
		t.Error(err)
	}

	// Check history for our bin
	req, err := http.NewRequest("POST", "http://testing.geobin.io/api/1/history/"+binId, nil)
	if err != nil {
		t.Error(err)
	}
	w := httptest.NewRecorder()
	gbs.ServeHTTP(w, req)

	assertResponseOK(w, t)

	var history []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &history); err != nil {
		t.Error(err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(history))
	}

	geo := history[0]

	if payload != geo["body"].(string) {
		t.Errorf("Expected: %s\nGot: %s", payload, geo["body"].(string))
	}
}

func TestCountsWorksAsIntended(t *testing.T) {
	gbs := createGeobinServer()
	bins, expected := createBins(gbs, []int{1, 0, 5, 23}, t)

	verifyCounts(gbs, bins, expected, t)
}

func TestCountsWithInvalidBinId(t *testing.T) {
	gbs := createGeobinServer()
	bins, expected := createBins(gbs, []int{1, 0, 2}, t)

	bins = append(bins, "invalid")
	expected["invalid"] = nil

	verifyCounts(gbs, bins, expected, t)
}

func TestRateLimitMiddleware(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}

	limited := createGeobinServer().rateLimit(handler, 1)

	req, _ := http.NewRequest("GET", "http://testing.geobin.io/", nil)
	wOk := httptest.NewRecorder()
	wErr := httptest.NewRecorder()

	// Rate limit is 1/s, first one should get a 200
	limited(wOk, req)
	// assertResponseOK(wOk, t)

	// Second request should get a 500
	limited(wErr, req)
	assertResponseCode(wErr, 500, t)
}

func createBins(gbs *geobinServer, counts []int, t *testing.T) (binIds []string, expected map[string]interface{}) {
	binIds = make([]string, len(counts))
	expected = make(map[string]interface{})

	// create the bins
	for i, c := range counts {
		binId, err := createBin(gbs)
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
			gbs.ServeHTTP(w, req)
			assertResponseOK(w, t)
		}
	}
	return
}

func verifyCounts(gbs *geobinServer, bins []string, expected map[string]interface{}, t *testing.T) {
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
	gbs.ServeHTTP(w, req)
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

func createGeobinServer() *geobinServer {
	ps := &MockPubSub{}
	return NewGeobinServer(testConf, NewMockRedis(), ps, NewSocketMap(ps))
}

func createBin(gbs *geobinServer) (string, error) {
	req, err := http.NewRequest("POST", "http://testing.geobin.io/api/1/create", nil)
	if err != nil {
		return "", err
	}
	w := httptest.NewRecorder()
	gbs.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		return "", errors.New(fmt.Sprint("Error creating bin: ", w.Body))
	}

	var js map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &js)
	id, ok := js["id"]
	if !ok {
		return "", errors.New("Invalid response from /create")
	}

	return id.(string), nil
}

func postToBin(gbs *geobinServer, binId string, payload string) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest("POST", "http://testing.geobin.io/"+binId, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	gbs.ServeHTTP(w, req)

	return w, nil
}
