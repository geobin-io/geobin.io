##### ⚠️ DEPRECATION NOTICE ⚠️

**This project is no longer maintained.**

---

# Geobin

_Inspect HTTP Requests with Geographic Data._

Geobin allows you to create a temporary URL to collect and inspect JSON requests that contain geographic data. Inspect the headers and body of a request and view any geographic data parsed out of the body on a map.

You can also receive data in real time. As long as your browser supports [WebSockets], once you're looking at a geobin, new data will be streamed directly to the browser and any found geographic data will be mapped instantly.

Hat tip to [RequestBin] for inspiration.

## Contents

* [Documentation]
* [Running Geobin locally]
* [How do we find geographic data?]
* [License]

## Documentation

* [Server]
* [Client]
* [API]

## Running Geobin locally

Requirements:

* [go]
* [node]
* [redis]

Here is the short version of how to get Geobin up and running locally, assuming you have a functional [go] environment, [node] environment, and [redis] server already set up on your machine.

### 1. `go get` Geobin

```bash
> go get github.com/esripdx/geobin.io
```

### 2. Run the setup scripts

```bash
> cd $GOPATH/src/github.com/esripdx/geobin.io
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

## License

Copyright 2014 Esri, Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

> http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

A copy of the license is available in the repository's [LICENSE.txt] file.

[Documentation]: #documentation
[Running Geobin locally]: #running-geobin-locally
[How do we find geographic data?]: #how-do-we-find-geographic-data
[GeoJSON]: http://geojson.org/geojson-spec.html
[WebSockets]: http://caniuse.com/websockets
[RequestBin]: http://requestb.in
[go]: http://golang.org
[Server]: static/doc/server.md
[Client]: static/doc/client.md
[API]: static/doc/api.md
[redis]: http://redis.io
[node]: http://nodejs.org
[License]: #license
[LICENSE.txt]: LICENSE.txt
