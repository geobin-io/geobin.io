package main

import (
	"testing"

	"encoding/json"
	"reflect"
	"strings"

	"github.com/kr/pretty"
)

var r = strings.NewReplacer(" ", "", "\n", "", "\t", "")

func init() {
	// make the default for isDebug be true when running tests. If you run `go test -debug=false`
	// the tests will not print out the debug info.
	*isDebug = true
}

func runSingleObjectTest(src string, t *testing.T) {
	// make expected string
	var expected map[string]interface{}
	if err := json.Unmarshal([]byte(src), &expected); err != nil {
		t.Error(err)
		return
	}
	expected["geobinRequestPath"] = make([]interface{}, 0)
	expSlice := []map[string]interface{}{expected}
	exp, _ := json.Marshal(expSlice)

	runTest(src, string(exp), t)
}

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

func TestRequestWithNonGJPoint(t *testing.T) {
	src := `{
		"foo": "bar",
		"lat": 10,
		"lng": -10
	}`
	exp := `[{
		"type": "Point",
		"coordinates": [-10, 10],
		"geobinRequestPath": []
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
		"type": "Point",
		"coordinates": [-10, 10],
		"geobinRequestPath": [0]
	}, {
		"type": "Point",
		"coordinates": [-20, 20],
		"geobinRequestPath": [1]
	}]`
	runTest(src, exp, t)
}

func TestRequestwithNonGJPointAndRadius(t *testing.T) {
	src := `{
		"foo": "bar",
		"lat": 10,
		"lng": -10,
		"rad": 10
	}`
	exp := `[{
		"type": "Point",
		"coordinates": [-10, 10],
		"geobinRequestPath": [],
		"geobinRadius": 10
	}]`
	runTest(src, exp, t)
}

func TestGTCallbackRequest(t *testing.T) {
	// TODO: use gtCallback.json file!
}
