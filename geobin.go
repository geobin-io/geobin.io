package main

import (
	"encoding/json"
	"fmt"
	gtjson "github.com/Esri/geotrigger-go/geotrigger/json"
	"github.com/geoloqi/geobin-go/manager"
	"github.com/geoloqi/geobin-go/socket"
	"github.com/gorilla/mux"
	gj "github.com/kpawlik/geojson"
	gu "github.com/nu7hatch/gouuid"
	redis "github.com/vmihailenco/redis/v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host       string
	Port       int
	RedisHost  string
	RedisPass  string
	RedisDB    int64
	NameVals   string
	NameLength int
}

var config = &Config{}
var client = &redis.Client{}
var pubsub = &redis.PubSub{}
var socketManager = manager.NewManager(make(map[string]map[string]socket.S))

type GeobinRequest struct {
	Timestamp int64                  `json:"timestamp"`
	Headers   map[string]string      `json:"headers"`
	Body      string                 `json:"body"`
	Geo       map[string]interface{} `json:"geo,omitempty"`
}

func NewGeobinRequest(ts int64, h map[string]string, b []byte) *GeobinRequest {
	gr := GeobinRequest{
		Timestamp: ts,
		Headers:   h,
		Body:      string(b),
	}

	js, foundGeojson := gr.parseGeojson()
	// TODO: use js in the !foundGeojson block below (it is the raw json found in the body, if any)
	_ = js

	// If we didn't find any geojson search for any coordinates in the body.
	if !foundGeojson {
		// TODO: Look for Lat/Lng (and Distance) keys, create geojson Features for each of them
		// placing any additional data near those keys in the properties key.
		// The code below is just my initial lat/long detection code: needs improvement
		latRegex := regexp.MustCompile(`.*(?:lat(?:itude)?|y)(?:")*: ?([0-9.-]*)`)
		lngRegex := regexp.MustCompile(`.*(?:lo?ng(?:itude)?|x)(?:")*: ?([0-9.-]*)`)

		var lat, lng float64
		var foundLat, foundLng bool
		bStr := string(b)
		if latMatches := latRegex.FindStringSubmatch(bStr); latMatches != nil {
			lat, _ = strconv.ParseFloat(latMatches[1], 64)
			foundLat = true
		}

		if lngMatches := lngRegex.FindStringSubmatch(bStr); lngMatches != nil {
			lng, _ = strconv.ParseFloat(lngMatches[1], 64)
			foundLng = true
		}

		if foundLat && foundLng {
			p := gj.NewPoint(gj.Coordinate{gj.CoordType(lng), gj.CoordType(lat)})
			pstr, _ := gj.Marshal(p)
			json.Unmarshal([]byte(pstr), &gr.Geo)
		}
	}

	if gr.Geo != nil {
		fmt.Fprintln(os.Stdout, "Found geo:", gr.Geo)
	} else {
		fmt.Fprintln(os.Stdout, "No geo found in request:", gr.Body)
	}
	return &gr
}

func (gr *GeobinRequest) parseGeojson() (js map[string]interface{}, foundGeojson bool) {
	var t string
	var geo interface{}
	b := []byte(gr.Body)

	if err := json.Unmarshal(b, &js); err != nil {
		fmt.Fprintln(os.Stdout, "error unmarshalling json:", err)
	} else {
		fmt.Fprintln(os.Stdout, "Json unmarshalled:", js)
		if err := gtjson.GetValueFromJSONObject(js, "type", &t); err != nil {
			fmt.Fprintln(os.Stdout, "Get value from json for type failed:", err)
		} else {
			unmarshal := func(buf []byte, target interface{}) (e error) {
				if e = json.Unmarshal(buf, target); e != nil {
					fmt.Fprintf(os.Stdout, "couldn't unmarshal %v to geojson: %v\n", t, e)
				}
				return
			}

			fmt.Fprintln(os.Stdout, "Found type:", t)
			switch t {
			case "Point":
				var p gj.Point
				if err = unmarshal(b, &p); err != nil {
					break
				}

				geo = p
				foundGeojson = true
				break
			case "LineString":
				var ls gj.LineString
				if err = unmarshal(b, &ls); err != nil {
					break
				}

				geo = ls
				foundGeojson = true
				break
			case "Polygon":
				var p gj.Polygon
				if err = unmarshal(b, &p); err != nil {
					break
				}

				geo = p
				foundGeojson = true
				break
			case "MultiPoint":
				var mp gj.MultiPoint
				if err = unmarshal(b, &mp); err != nil {
					break
				}

				geo = mp
				foundGeojson = true
				break
			case "MultiPolygon":
				var mp gj.MultiPolygon
				if err = unmarshal(b, &mp); err != nil {
					break
				}

				geo = mp
				foundGeojson = true
				break
			case "GeometryCollection":
				var gc gj.GeometryCollection
				if err = unmarshal(b, &gc); err != nil {
					break
				}

				geo = gc
				foundGeojson = true
				break
			case "Feature":
				var f gj.Feature
				if err = unmarshal(b, &f); err != nil {
					break
				}

				geo = f
				foundGeojson = true
				break
			case "FeatureCollection":
				var fc gj.FeatureCollection
				if err = unmarshal(b, &fc); err != nil {
					break
				}

				geo = fc
				foundGeojson = true
				break
			default:
				fmt.Fprintln(os.Stdout, "Unknown geo type:", t)
				break
			}
		}
	}
	if foundGeojson {
		gjs, _ := gj.Marshal(geo)
		json.Unmarshal([]byte(gjs), &gr.Geo)
	}
	return js, foundGeojson
}

func main() {
	// loop for receiving messages from Redis pubsub, and forwarding them on to relevant ws connection
	go redisPump()

	defer func() {
		pubsub.Close()
		client.Close()
	}()

	// Start up HTTP server
	fmt.Fprintf(os.Stdout, "Starting server at %v:%d\n", config.Host, config.Port)
	err := http.ListenAndServe(fmt.Sprintf("%v:%d", config.Host, config.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

/*
 * Initilization
 */
func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// add file info to log statements
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	// set up unique seed for random num generation
	rand.Seed(time.Now().UTC().UnixNano())

	// prepare router
	r := createRouter()
	http.Handle("/", r)

	loadConfig()
	setupRedis()
}

func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
}

func setupRedis() {
	client = redis.NewTCPClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})

	if ping := client.Ping(); ping.Err() != nil {
		log.Fatal(ping.Err())
	}
	pubsub = client.PubSub()
}

func createRouter() *mux.Router {
	r := mux.NewRouter()
	// API routes (POSTs only!)
	api := r.Methods("POST").PathPrefix("/api/{v:[0-9.]+}/").Subrouter()
	api.HandleFunc("/create", createHandler)
	api.HandleFunc("/history/{name}", historyHandler)
	api.HandleFunc("/ws/{name}", wsHandler)

	// Our bread and/or butter (how requests actually get put into redis)
	r.HandleFunc("/{name}", binHandler).Methods("POST")

	// Client/web requests (GETs only!)
	web := r.Methods("GET").Subrouter()
	// Any GET request to the /api/ route will serve up the docs static site directly.
	web.PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "docs - %v\n", req.URL)
		// TODO: This is wrong, will fix when we actually have the files to serve
		http.ServeFile(w, req, "docs/build/")
	})
	// Any GET request to the /static/ route will serve the files in the static dir directly.
	web.PathPrefix("/static/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "static - %v\n", req.URL)
		http.ServeFile(w, req, req.URL.Path[1:])
	})
	// All other GET requests will serve up the Angular app at static/index.html
	web.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(os.Stdout, "web - %v\n", req.URL)
		http.ServeFile(w, req, "static/index.html")
	})
	return r
}

/*
 * API Routes
 */
func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "create - %v\n", r.URL)

	// Get a new name
	n, err := randomString(config.NameLength)
	if err != nil {
		log.Println("Failure to create new name:", n, err)
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	// Save to redis
	if res := client.ZAdd(n, redis.Z{0, ""}); res.Err() != nil {
		log.Println("Failure to ZADD to", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}

	// Set expiration
	d := 48 * time.Hour
	if res := client.Expire(n, d); res.Err() != nil {
		log.Println("Failure to set EXPIRE for", n, res.Err())
		http.Error(w, "Could not generate new Geobin!", http.StatusInternalServerError)
		return
	}
	exp := time.Now().Add(d).Unix()

	// Create the json response and encoder
	encoder := json.NewEncoder(w)
	bin := map[string]interface{}{
		"id":      n,
		"expires": exp,
	}

	// encode the json directly to the response writer
	err = encoder.Encode(bin)
	if err != nil {
		log.Println("Failure to create json for new name:", n, err)
		http.Error(w, fmt.Sprintf("New Geobin created (%v) but we could not return the JSON for it!", n), http.StatusInternalServerError)
		return
	}
}

func binHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "bin - %v\n", r.URL)
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

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error while reading POST body:", err)
		http.Error(w, "Could not read POST body!", http.StatusInternalServerError)
		return
	}

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[k] = strings.Join(v, ", ")
	}

	gr := NewGeobinRequest(time.Now().UTC().Unix(), headers, body)

	encoded, err := json.Marshal(gr)
	if err != nil {
		log.Println("Error marshalling request:", err)
	}

	if res := client.ZAdd(name, redis.Z{float64(time.Now().UTC().Unix()), string(encoded)}); res.Err() != nil {
		log.Println("Failure to ZADD to", name, res.Err())
	}

	if res := client.Publish(name, string(encoded)); res.Err() != nil {
		log.Println("Failure to PUBLISH to", name, res.Err())
	}
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "history - %v\n", r.URL)
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

	set := client.ZRevRange(name, "0", "-1")
	if set.Err() != nil {
		log.Println("Failure to ZREVRANGE for", name, set.Err())
	}

	// chop off the last history member since it is the placeholder value from when the set was created
	vals := set.Val()[:len(set.Val())-1]

	history := make([]GeobinRequest, 0, len(vals))
	for _, v := range vals {
		var gr GeobinRequest
		if err := json.Unmarshal([]byte(v), &gr); err != nil {
			log.Println("Error unmarshalling request history:", err)
		}
		history = append(history, gr)
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(history)
	if err != nil {
		log.Println("Error marshalling request history:", err)
		http.Error(w, "Could not generate history.", http.StatusInternalServerError)
		return
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "create - %v\n", r.URL)
	// upgrade the connection
	binName := mux.Vars(r)["name"]

	// start pub subbing
	if err := pubsub.Subscribe(binName); err != nil {
		log.Println("Failure to SUBSCRIBE to", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := gu.NewV4()
	if err != nil {
		log.Println("Failure to generate new socket UUID", binName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	uuid := id.String()

	s, err := socket.NewSocket(binName+"~br~"+uuid, w, r, nil, func(socketName string) {
		// the socketname is a composite of the bin name, and the socket UUID
		ids := strings.Split(socketName, "~br~")
		bn := ids[0]
		suuid := ids[1]

		manageSockets(func(sockets map[string]map[string]socket.S) {
			socks, ok := sockets[bn]
			if ok {
				delete(socks, suuid)

				if len(socks) == 0 {
					delete(sockets, bn)
					if err := pubsub.Unsubscribe(bn); err != nil {
						log.Println("Failure to UNSUBSCRIBE from", bn, err)
					}
				}
			}
		})
	})

	if err != nil {
		// if there is an error, NewSocket will have already written a response via http.Error()
		// so only write a log
		log.Println("Error opening websocket:", err)
		return
	}

	// keep track of the outbound channel for pubsubbery
	go manageSockets(func(sockets map[string]map[string]socket.S) {
		if _, ok := sockets[binName]; !ok {
			sockets[binName] = make(map[string]socket.S)
		}
		sockets[binName][uuid] = s
	})
}

/*
 * Redis
 */
func redisPump() {
	for {
		v, err := pubsub.Receive()
		if err != nil {
			log.Println("Error from Redis PubSub:", err)
			return
		}

		switch v := v.(type) {
		case *redis.Message:
			var sockMap map[string]socket.S
			var ok bool
			manageSockets(func(sockets map[string]map[string]socket.S) {
				sockMap, ok = sockets[v.Channel]
			})

			if !ok {
				log.Println("Got message for unknown channel:", v.Channel)
				return
			}

			for _, sock := range sockMap {
				go func(s socket.S, p []byte) {
					s.Write(p)
				}(sock, []byte(v.Payload))
			}
		}
	}
}

/*
 * Utils
 */
func randomString(length int) (string, error) {
	b := make([]byte, length)
	for i, _ := range b {
		b[i] = config.NameVals[rand.Intn(len(config.NameVals))]
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
	}

	return resp.Val(), nil
}

func manageSockets(sf func(sockets map[string]map[string]socket.S)) {
	socketManager.Touch(func(o interface{}) {
		if sockets, ok := o.(map[string]map[string]socket.S); ok {
			sf(sockets)
		}
	})
}
