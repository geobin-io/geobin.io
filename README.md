# Geobin

_It's like [RequestBin], but different._

Geobin allows you to create a temporary URL to collect and inspect JSON requests with geographic data. Inspect the headers and body of a request and view any geographic data parsed out of the body on a map.

You can also receive data in real time. As long as your browser supports [WebSockets], once you're looking at a geobin, new data will be streamed directly to the browser and any found geographic data will be mapped instantly.

## How do we find geographic data?

We look for [valid](http://geojsonlint.com) [GeoJSON]. If no GeoJSON is detected, we'll also look for the following properties:

### Latitude & Longitude

* expected format:

```javascript
{
  "latitude": 0,
  "longitude": 0
}
```

* accepted keys:
  * Latitude:
    * `lat`
    * `latitude`
    * `y`
  * Longitude:
    * `lng`
    * `lon`
    * `long`
    * `longitude`
    * `x`


### Radius (in meters)

* expected format:

```javascript
{
  "latitude": 0,
  "longitude": 0,
  "radius": 0
}
```

* accepted keys:
  * `dst`
  * `dist`
  * `distance`
  * `rad`
  * `radius`
  * `acc`
  * `accuracy`

### Coordinates

* expected format:

```javascript
{
  "coords": [0,0] // (x (longitude), y (latitude))
}
```

* accepted keys:
  * `geo`
  * `loc`
  * `location`
  * `coord`
  * `coordinate`
  * `coords`
  * `coordinates`

## Running Geobin locally

Here is the short version of how to get Geobin up and running locally, assuming you have a functional [go] environment and [redis] server already set up on your machine.

### 1. `go get` Geobin

```bash
> go get github.com/geoloqi/geobin-go
```

### 2. Run the setup scripts

```bash
> cd $GOPATH/src/github.com/geoloqi/geobin-go
> make setup # (runs `go get -t` and `npm install`)
```

### 3. Copy/Edit `config.json`

```bash
> cp config.json.dist config.json
> vim config.json # Optional, the defaults will work for running locally and connecting to a local redis on the default port
```

### 4. Run the server

```bash
> make run
```

You're up and running, have fun! Try opening http://localhost:8080 in a browser and clicking the "Create a New Geobin" button. Then run the following in a console while keeping an eye on your browser:

```bash
> curl -i -X POST \
-H "Content-Type: application/json" \
-d @gtCallback.json http://localhost:8080/BIN_ID
```

## Documentation

* [Server]
* [Client]
* [API]

[GeoJSON]: http://geojson.org/geojson-spec.html
[WebSockets]: http://caniuse.com/websockets
[RequestBin]: http://requestb.in
[go]: http://golang.org
[Server]: static/doc/server.md
[Client]: static/doc/client.md
[API]: static/doc/api.md
[redis]: http://redis.io
