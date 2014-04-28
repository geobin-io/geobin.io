package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

// NewGeobinRequest creates a new GeobinRequest with the given timestamp,
// headers, and body. It will search the given body for the presence of
// any geo data and fill the returned GeobinRequest's Geo property with
// an array of geoJSON objects using said geo data.
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
	go gr.parse(js, make([]interface{}, 0))
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

	return &gr
}

func (gr *GeobinRequest) parse(b interface{}, kp []interface{}) {
	switch t := b.(type) {
	case []interface{}:
		gr.parseArray(t, kp)
	case map[string]interface{}:
		gr.parseObject(t, kp)
	}
	gr.wg.Done()
}

func (gr *GeobinRequest) parseObject(o map[string]interface{}, kp []interface{}) {
	if isGeojson(o) {
		o["geobinRequestPath"] = kp
		gr.c <- o
	} else if foundGeo, geo := isOtherGeo(o); foundGeo {
		geo["geobinRequestPath"] = kp
		gr.c <- geo
	} else {
		for k, v := range o {
			kp = append(kp, k)
			gr.wg.Add(1)
			go gr.parse(v, kp)
		}
	}
}

func (gr *GeobinRequest) parseArray(a []interface{}, kp []interface{}) {
	for i, o := range a {
		kp = append(kp, i)
		gr.wg.Add(1)
		go gr.parse(o, kp)
	}
}

// isOtherGeo searches for non-standard geo data in the given json map. It looks for the presence
// of lat/lng (and a few variations thereof) or x/y values in the object as well as a distance/radius/accuracy
// field and creates a geojson point out of it and returns that, along with a boolean value
// representing whether or not it found any geo data in the object. It will also look for
// any keys that hold an array of two numbers with a key name that suggests that it might
// be a lng/lat array.
//
// The following keys will be detected as Latitude:
//	"lat", "latitude"
//	"y"
//
// The following keys will be detected as Longitude:
//	"lng", "long", "longitude"
//	"x"
//
// The following keys will be used to fill the "geobinRadius" property of the resulting geojson:
//	"dist", "distance"
//	"rad", "radius"
//	"acc", "accuracy"
//
// The following keys will be searched for a long/lat pair:
//	"geo"
//	"loc" or "location"
//	"coord", "coords", "coordinate" or "coordinates"
func isOtherGeo(o map[string]interface{}) (bool, map[string]interface{}) {
	var foundLat, foundLng, foundDst bool
	var lat, lng, dst float64

	for k, v := range o {
		switch strings.ToLower(k) {
		case "lat", "latitude", "y":
			lat, foundLat = v.(float64)
		case "lng", "long", "longitude", "x":
			lng, foundLng = v.(float64)
		case "dst", "dist", "distance", "rad", "radius", "acc", "accuracy":
			dst, foundDst = v.(float64)
		case "geo", "loc", "location", "coord", "coordinate", "coords", "coordinates":
			g, ok := v.([]float64)
			if !ok || len(g) != 2 {
				break
			}

			lng = g[0]
			lat = g[1]
			foundLat, foundLng = true, true
		}
	}

	if foundLat && foundLng {
		p := gj.NewPoint(gj.Coordinate{gj.CoordType(lng), gj.CoordType(lat)})
		pstr, _ := gj.Marshal(p)
		var geo map[string]interface{}
		json.Unmarshal([]byte(pstr), &geo)
		if foundDst {
			geo["geobinRadius"] = dst
		}
		fmt.Fprintln(os.Stdout, "Found other geo:", geo)
		return true, geo
	}

	return false, nil
}

// isGeojson detects whether or not the given json map is valid GeoJSON and
// returns a boolean reflecting its findings.
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
