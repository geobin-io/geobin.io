package main

import "testing"
import "github.com/geoloqi/geobin-go/test"
import "strings"

func TestRequestWithGJPoint(t *testing.T) {
	js := `{ "type": "Point", "coordinates": [100, 0] }`

	gr := NewGeobinRequest(0, nil, []byte(js))
	test.Expect(t, gr.Geo, strings.Replace(js, " ", "", -1))
}

func TestRequestWithGJLineString(t *testing.T) {
	js := `{ "type": "LineString", "coordinates": [ [100, 0], [101, 1] ] }`

	gr := NewGeobinRequest(0, nil, []byte(js))
	test.Expect(t, gr.Geo, strings.Replace(js, " ", "", -1))
}

/*
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
}

func TestRequestWithGJMultiPoint(t *testing.T) {
	js := `{
		"type": "MultiPoint",
		"coordinates": [ [100, 0], [101, 1] ]
	}`
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
}

func TestRequestWithGJGeometryCollection(t *testing.T) {
	js := `{
		"type": "GeometryCollection",
    "geometries": [
      {
				"type": "Point",
        "coordinates": [100, 0]
			},
      {
				"type": "LineString",
        "coordinates": [ [101, 0], [102, 1] ]
			}
    ]
  }`
}

func TestRequestWithGJFeatur(t *testing.T) {
	// TODO:
}

func TestRequestWithGJFeatureCollection(t *testing.T) {
	// TODO:
}

func TestRequestWithNoGJPoint(t *testing.T) {
	// TODO:
}

func TestRequestWithNoGJPoints(t *testing.T) {
	// TODO:
}

func TestRequestwithNoGJPointAndRadius(t *testing.T) {
	// TODO:
}
*/
