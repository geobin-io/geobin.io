package main

import "testing"
import "encoding/json"
import "strings"
import "reflect"

var r *strings.Replacer = strings.NewReplacer(" ", "", "\n", "", "\t", "")

func runTest(js string, t *testing.T) {
	gr := NewGeobinRequest(0, nil, []byte(js))
	var jsMap map[string]interface{}
	if err := json.Unmarshal([]byte(js), &jsMap); err != nil {
		t.Error(err)
		return
	}
	if !reflect.DeepEqual(jsMap, gr.Geo) {
		t.Errorf("Expected %v, (type %v) - Got %v (type %v)", jsMap, reflect.TypeOf(jsMap), gr.Geo, reflect.TypeOf(gr.Geo))
		return
	}
}

func TestRequestWithGJPoint(t *testing.T) {
	runTest(`{ "type": "Point", "coordinates": [100, 0] }`, t)
}

func TestRequestWithGJLineString(t *testing.T) {
	runTest(`{ "type": "LineString", "coordinates": [ [100, 0], [101, 1] ] }`, t)
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

	runTest(jsNoHoles, t)
	runTest(jsHoles, t)
}

func TestRequestWithGJMultiPoint(t *testing.T) {
	js := `{
		"type": "MultiPoint",
		"coordinates": [ [100, 0], [101, 1] ]
	}`
	runTest(js, t)
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
	runTest(js, t)
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
	runTest(js, t)
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
	runTest(js, t)
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
	runTest(js, t)
}

func TestRequestWithNonGJPoint(t *testing.T) {
	// TODO:
}

func TestRequestWithNonGJPoints(t *testing.T) {
	// TODO:
}

func TestRequestwithNonGJPointAndRadius(t *testing.T) {
	// TODO:
}
