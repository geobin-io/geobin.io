package main

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/kr/pretty"
)

func init() {
	// make the default for isDebug be true when running tests. If you run `go test -debug=false`
	// the tests will not print out the debug info.
	*isDebug = true
}

// runTest takes an input json string (src) and the expected output string (expected). It creates
// a GeobinRequest using the input string and tests that the GeobinRequest's Geo property,
// once marshaled to json is the same as the result of marshaling the `expected` json string.
// Note that expected should be marshalable to []interface{}, as that is what GeobinRequest.Geo
// will be.
func runTest(src string, expected string, t *testing.T) {
	var exp []interface{}
	if err := json.Unmarshal([]byte(expected), &exp); err != nil {
		t.Error(err)
		return
	}

	gr := NewGeobinRequest(0, nil, []byte(src))

	// convert gr.Geo to json, and back to avoid funky type differences (int vs float) and to
	// test gr.Geo as it will be seen by the client.
	var res []interface{}
	resB, _ := json.Marshal(gr.Geo)
	json.Unmarshal(resB, &res)

	if !reflect.DeepEqual(exp, res) {
		pretty.Logf("Expected:\n%# v,\nGot:\n%# v", exp, res)
		t.Fail()
		return
	}
}

// runSingleObjectTest is a wrapper function for runTest. It takes an input json string and
// generates the expected output parameter for runTest by unmarshaling the input and setting
// the geobinRequestPath as expected for an object that is the root of the request body.
func runSingleObjectTest(src string, t *testing.T) {
	// make expected geo
	var geo map[string]interface{}
	if err := json.Unmarshal([]byte(src), &geo); err != nil {
		t.Error(err)
		return
	}

	// nest it inside Geo container, wrap that in a slice and marshal it to json
	expected := make(map[string]interface{})
	expected["geo"] = geo
	expected["path"] = make([]interface{}, 0)
	expSlice := []map[string]interface{}{expected}
	exp, _ := json.Marshal(expSlice)

	runTest(src, string(exp), t)
}

// GeoJSON Tests

func TestRequestWithGJPoint(t *testing.T) {
	runSingleObjectTest(`{ "type": "Point", "coordinates": [100, 0] }`, t)
}

func TestRequestWithGJLineString(t *testing.T) {
	runSingleObjectTest(`{ "type": "LineString", "coordinates": [ [100, 0], [101, 1] ] }`, t)
}

func TestRequestWithGJPolygon(t *testing.T) {
	jsNoHoles := `{
		"type": "Polygon",
    "coordinates": [
      [ [100, 0], [101, 0], [101, 1], [100, 1], [100, 0] ]
		]
	}`
	jsHoles := `{
		"type": "Polygon",
		"coordinates": [
			[ [100, 0], [101, 0], [101, 1], [100, 1], [100, 0] ],
			[ [100.2, 0.2], [100.8, 0.2], [100.8, 0.8], [100.2, 0.8], [100.2, 0.2] ]
		]
	}`

	runSingleObjectTest(jsNoHoles, t)
	runSingleObjectTest(jsHoles, t)
}

func TestRequestWithGJMultiPoint(t *testing.T) {
	js := `{
		"type": "MultiPoint",
		"coordinates": [ [100, 0], [101, 1] ]
	}`
	runSingleObjectTest(js, t)
}

func TestRequestWithGJMultiPolygon(t *testing.T) {
	js := `{
		"type": "MultiPolygon",
    "coordinates": [
      [[[102, 2], [103, 2], [103, 3], [102, 3], [102, 2]]],
      [[[100, 0], [101, 0], [101, 1], [100, 1], [100, 0]],
			[[100.2, 0.2], [100.8, 0.2], [100.8, 0.8], [100.2, 0.8], [100.2, 0.2]]]
		]
	}`
	runSingleObjectTest(js, t)
}

func TestRequestWithGJGeometryCollection(t *testing.T) {
	js := `{
		"type": "GeometryCollection",
		"geometries": [
		{
			"coordinates": [100, 0],
					"type": "Point"
				},
		{
			"coordinates": [ [101, 0], [102, 1] ],
					"type": "LineString"
				}
		]
	}`
	runSingleObjectTest(js, t)
}

func TestRequestWithGJFeature(t *testing.T) {
	js := `{
		"type": "Feature",
		"id": "feature-test",
		"geometry": {
			"coordinates": [-122.65, 45.51],
			"type": "Point"
		},
		"properties": {
			"foo": "bar"
		}
	}`
	runSingleObjectTest(js, t)
}

func TestRequestWithGJFeatureCollection(t *testing.T) {
	js := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"id": "feature-test",
				"geometry": {
					"coordinates": [-122.65, 45.51],
					"type": "Point"
				},
				"properties": {
					"foo": "bar"
				}
			}
		]
	}`
	runSingleObjectTest(js, t)
}

func TestRequestWithNestedGeoJSON(t *testing.T) {
	src := `{
		"foo": "bar",
		"data": {
			"foo": "baz",
			"properties": {
				"geo": {
					"type": "Point",
					"coordinates": [10, -10]
				},
				"someOtherProperty": "with some other value"
			}
		}
	}`
	exp := `[{
		"geo": {
			"type": "Point",
			"coordinates": [10, -10]
		},
		"path": ["data", "properties", "geo"]
	}]`
	runTest(src, exp, t)
}

// Other Geo Tests

func TestRequestWithNonGJPoint(t *testing.T) {
	src := `{
		"foo": "bar",
		"lat": 10,
		"lng": -10
	}`
	exp := `[{
		"geo": {
			"type": "Point",
			"coordinates": [-10, 10]
		},
		"path": []
	}]`
	runTest(src, exp, t)
}

func TestRequestWithNonGJPoints(t *testing.T) {
	src := `[{
		"foo": "bar",
		"lat": 10,
		"lng": -10
	}, {
		"foo": "baz",
		"x": -20,
		"y": 20
	}]`
	exp := `[{
		"geo": {
			"type": "Point",
			"coordinates": [-10, 10]
		},
		"path": [0]
	}, {
		"geo": {
			"type": "Point",
			"coordinates": [-20, 20]
		},
		"path": [1]
	}]`
	runTest(src, exp, t)
}

// Ensure that we find lat and long values for all of the variations of the
// key name
func TestRequestWithNonGJLatLngKeys(t *testing.T) {
	latKeys := []string{"lat", "latitude", "y"}
	lngKeys := []string{"lng", "lon", "long", "longitude", "x"}

	// For every combination of latKeys and lngKeys, we expect the same result
	exp := `[{"geo": {"type": "Point", "coordinates": [-10, 10]}, "path": []}]`
	for _, latKey := range latKeys {
		for _, lngKey := range lngKeys {
			src := `{"` + latKey + `": 10, "` + lngKey + `": -10}`
			runTest(src, exp, t)
		}
	}
}

// Ensure that we find dist values for all variations of the key name
func TestRequestWithNonGJDistKeys(t *testing.T) {
	distKeys := []string{"dst", "dist", "distance", "rad", "radius", "acc", "accuracy"}

	// For each distKey, we expect the same result
	exp := `[{"geo": {"type": "Point", "coordinates": [-10, 10]}, "radius": 5, "path": []}]`
	for _, distKey := range distKeys {
		src := `{"lat": 10, "lng": -10, "` + distKey + `": 5}`
		runTest(src, exp, t)
	}

}

// Ensure that we find geo objects for all variations of the key name.
// Also ensures that the geobinRequestPath is correct in all cases.
func TestRequestWithNonGJGeoKeys(t *testing.T) {
	geoKeys := []string{"geo", "loc", "location", "coord", "coordinate", "coords", "coordinates"}

	// For each geoKey, we expect the same coordinates, with geobinRequestPath = [geoKey]
	for _, geoKey := range geoKeys {
		exp := `[{"geo": {"type": "Point", "coordinates": [-10, 10]}, "path": ["` + geoKey + `"]}]`
		src := `{"` + geoKey + `": {"lat": 10, "lng": -10}}`
		runTest(src, exp, t)
	}
}

func TestRequestwithNonGJPointAndRadius(t *testing.T) {
	src := `{
		"foo": "bar",
		"lat": 10,
		"lng": -10,
		"rad": 10
	}`
	exp := `[{
		"geo": {
			"type": "Point",
			"coordinates": [-10, 10]
		},
		"path": [],
		"radius": 10
	}]`
	runTest(src, exp, t)
}

func TestGTCallbackRequest(t *testing.T) {
	js, err := ioutil.ReadFile("gtCallback.json")
	if err != nil {
		t.Error("Error reading gtCallback.json. ", err)
		return
	}

	expJs := `[
		{
			"geo": {
				"type": "Point",
				"coordinates": [-122.67545711249113, 45.51986460661744]
			},
			"radius": 8,
			"path": ["location"]
		},
		{
			"geo": {
				"type": "Point",
				"coordinates": [-122.77545711249113, 45.41986460661744]
			},
			"path": ["trigger", "condition", "geo"]
		}
	]`

	runTest(string(js), expJs, t)
}
