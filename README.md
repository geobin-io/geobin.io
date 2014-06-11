# Geobin

_It's like [RequestBin], but different._

Geobin allows you to create a temporary URL to collect and inspect JSON requests with geographic data. Inspect the headers and body of a request and view any geographic data parsed out of the body on a map.

You can also receive data in real time. As long as your browser supports [WebSockets], once you're looking at a geobin, new data will be streamed directly to the browser and any found geographic data will be mapped instantly.

## How do we find geographic data?

We look for [valid](http://geojsonlint.com) [GeoJSON]. If no GeoJSON is detected, we'll also look for the following properties:

<ul>
<li>__Latitude &amp; Longitude__

_expected format:_

```javascript
{
  "latitude": 0,
  "longitude": 0
}
```

_accepted keys:_

* __Latitude__: 

	* `lat`
	* `latitude`
	* `y`

* __Longitude__:

	* `lng`
	* `lon`
	* `long`
	* `longitude`
	* `x`
</li>

<li>__Radius__ (in meters)

_expected format:_

```javascript
{
  "latitude": 0,
  "longitude": 0,
  "radius": 0
}
```

_accepted keys:_

* `dst`
* `dist`
* `distance`
* `rad`
* `radius`
* `acc`
* `accuracy`
</li>

<li>__Coordinates__ 

_expected format:_

```javascript
{
  "coords": [0,0] // (x (longitude), y (latitude))
}
```

_accepted keys:_

* `geo`
* `loc`
* `location`
* `coord`
* `coordinate`
* `coords`
* `coordinates`

</li>
</ul>

## Running Geobin locally

Here is the short version of how to get Geobin up and running locally, assuming you have a functional [go] environment already set up on your machine.

<ol>
<li>`go get` Geobin.

```bash
> go get github.com/geoloqi/geobin-go
```
</li>

<li>Run the setup scripts.

```bash
> cd $GOPATH/src/github.com/geoloqi/geobin-go
> make setup # (runs `go get -t` and `npm install`)
```
</li>

<li>Run the server

```bash
> make run
```
</li>

<li>You're up and running, have fun! Try opening http://localhost:8080 in a browser and clicking the "Create a New Geobin" button. Then run the following in a console while keeping an eye on your browser:
```bash
> curl -i -X POST -H "Content-Type: application/json" -d @gtCallback.json http://localhost:8080/BIN_ID`
```
</li>

For more details, see the [server], [client], and/or [API] docs.

[GeoJSON]: http://geojson.org/geojson-spec.html 
[WebSockets]: http://caniuse.com/websockets
[RequestBin]: http://requestb.in
[go]: http://golang.org
[server]: doc/server.md
[client]: doc/client.md
[API]: doc/api.md
