package main

import (
	gtjson "github.com/Esri/geotrigger-go/geotrigger/json"
	gj "github.com/kpawlik/geojson"
	"regexp"
	"strconv"
	"encoding/json"
	"fmt"
	"os"
)

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
