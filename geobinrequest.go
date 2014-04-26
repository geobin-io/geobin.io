package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	gj "github.com/kpawlik/geojson"
)

// GeobinRequest stores received data and any detected geo info from a request
type GeobinRequest struct {
	Timestamp int64             `json:"timestamp"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
	Geo       []interface{}     `json:"geo,omitempty"`
	wg        *sync.WaitGroup
	c         chan map[string]interface{}
}

// NewGeobinRequest creates and new GeobinRequest with the given timestamp,
// headers, and body
func NewGeobinRequest(timestamp int64, headers map[string]string, body []byte) *GeobinRequest {
	gr := GeobinRequest{
		Timestamp: timestamp,
		Headers:   headers,
		Body:      string(body),
		wg:        &sync.WaitGroup{},
		c:         make(chan map[string]interface{}),
	}

	var js interface{}
	if err := json.Unmarshal(body, &js); err != nil {
		fmt.Fprintln(os.Stdout, "No json found in request:", gr.Body)
		return &gr
	}

	gr.wg.Add(1)
	go gr.parse(js)
	go func() {
		for {
			geo, ok := <-gr.c
			if !ok {
				return
			}

			gr.Geo = append(gr.Geo, geo)
		}
	}()
	gr.wg.Wait()
	close(gr.c)

	/*
		// If we didn't find any geojson search for any coordinates in the body.
		if false {
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
	*/

	return &gr
}

func (gr *GeobinRequest) parse(b interface{}) {
	switch t := b.(type) {
	case []interface{}:
		gr.parseArray(t)
	case map[string]interface{}:
		gr.parseObject(t)
	}
	gr.wg.Done()
}

func (gr *GeobinRequest) parseObject(o map[string]interface{}) {
	if isGeojson(o) {
		gr.c <- o
	} else if isOtherGeo(o) {
	} else {
		for _, v := range o {
			gr.wg.Add(1)
			go gr.parse(v)
		}
	}
}

func isOtherGeo(o map[string]interface{}) bool {
	return false
}

func isGeojson(js map[string]interface{}) bool {
	t, ok := js["type"]
	unmarshal := func(buf []byte, target interface{}) (e error) {
		if e = json.Unmarshal(buf, target); e != nil {
			fmt.Fprintf(os.Stdout, "couldn't unmarshal %v to geojson: %v\n", t, e)
		}
		return
	}

	if !ok {
		return false
	}

	fmt.Fprintln(os.Stdout, "Found type:", t)
	b, err := json.Marshal(js)
	if err != nil {
		return false
	}
	switch t {
	case "Point":
		var p gj.Point
		if err = unmarshal(b, &p); err != nil {
			return false
		}

		return true
	case "LineString":
		var ls gj.LineString
		if err = unmarshal(b, &ls); err != nil {
			return false
		}

		return true
	case "Polygon":
		var p gj.Polygon
		if err = unmarshal(b, &p); err != nil {
			return false
		}

		return true
	case "MultiPoint":
		var mp gj.MultiPoint
		if err = unmarshal(b, &mp); err != nil {
			return false
		}

		return true
	case "MultiPolygon":
		var mp gj.MultiPolygon
		if err = unmarshal(b, &mp); err != nil {
			return false
		}

		return true
	case "GeometryCollection":
		var gc gj.GeometryCollection
		if err = unmarshal(b, &gc); err != nil {
			return false
		}

		return true
	case "Feature":
		var f gj.Feature
		if err = unmarshal(b, &f); err != nil {
			return false
		}

		return true
	case "FeatureCollection":
		var fc gj.FeatureCollection
		if err = unmarshal(b, &fc); err != nil {
			return false
		}

		return true
	default:
		fmt.Fprintln(os.Stdout, "Unknown geo type:", t)
		return false
	}
}

func (gr *GeobinRequest) parseArray(a []interface{}) {
	for _, o := range a {
		gr.wg.Add(1)
		go gr.parse(o)
	}
}
