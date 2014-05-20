# Geobin API Documentation
To hit any of these endpoints you must send a POST request. All GET requests will be routed to the web server.

## /{bin_id}
POSTs to this endpoint to send data to the specified bin.

### Input
POST to this endpoint the data you'd like to have visualized. This can be any arbitrary JSON formatted data.
Geobin will process the posted JSON data and find any geo data it can and store what it found and where in
the database so that it can be visualized with the web frontend. 

It currently will detect geo data in the following formats:

* Any GeoJSON in the request body will be pulled directly out unmodified. We will try to find GeoJSON nested
  at any level of the object as well.
* Arbitrary JSON objects that meet the follwoing criteria will be turned into a GeoJSON Point and stored
  in the database:
	* Contain at least _one of each_ of the following keys:
		* "lat", "latitude", "y"
		* "lng", "lon", "long", "longitude", "x"
	* Contain one of the following keys that has an array of two numbers as its value:
		* "geo"
		* "loc" or "location"
		* "coord", "coords", "coordinate" or "coordinates"
	* If either of the above arbitrary JSON object types are found we will also search for the following keys
	and store the value with the GeoJSON Point that we create so that we can draw the point and radius as a
	circle on the map.
		* "rad" or "radius"
		* "dist" or "distance"
		* "acc" or "accuracy"

### Example

```sh
> curl -X POST http://localhost:8080/PF4C5zm67N -d '{"lat": 10, "lng": -10}' -i
HTTP/1.1 200 OK
Date: Mon, 19 May 2014 22:38:53 GMT
Content-Length: 0
Content-Type: text/plain; charset=utf-8
```

## /api/1/create
POST to this endpoint to create a new bin with a 48 hour expiration time and returns a json object with the following structure:

### Input
The POST to this endpoint should have an empty request body.

### Output

```javascript
{
  "id": {bin_id},
  "expires": {expiration_timestamp}
}
```

The expiration timestamp is in Unix time (milis).

### Example
```sh
> curl -X POST http://geobin.io/api/1/create
{"expires":1400706585,"id":"PF4C5zm67N"}
```

## /api/1/counts
POST to this endpoint with a list of binIDs to get a map of the given binIDs to the number of requests stored
in that bin. 

### Input
The POST to this endpoint should include a JSON array of binIDs to get counts for.

### Output
The response from this endpoint will be a map of binIDs to number of requests received by that binID:

```javascript
{
  "{bin_id}": {count}
}
```

If any of the `bin_id`s are not found in the database the value for that `bin_id` will be `null`.

### Example

```sh
> curl -X POST http://localhost:8080/api/1/counts -d '["PF4C5zm67N","foo"]'
{"PF4C5zm67N":0,"foo":null}
```

## /api/1/history/{bin_id}
POSTs to this route return all of the stored requests for the specified bin.

### Input
The POST to this endpoint should have an empty request body.

### Output
Each item in the returned array will have the following format:
```javascript
{
  "timestamp": {Unix timestamp in milis}, // when the payload was received
  "headers": {map of the original request headers},
  "body": {string representation of the original request body we received},
  "geo": {an array of objects with the following keys:
	"geo": {the geoJSON data that was found or created},
	"path": {an array of keys used to traverse the body json to get to this item}
  },
}
```

### Example
```sh
> curl -X POST http://localhost:8080/api/1/history/PF4C5zm67N
[ {
  "timestamp":1400539133,
	"headers":{
	  "Accept":"*/*",
	  "Content-Length":"23",
	  "Content-Type":"application/x-www-form-urlencoded",
	  "User-Agent":"curl/7.30.0"
	},
	"body": "{\"lat\": 10, \"lng\": -10}",
	"geo": [ {
	  "geo": { "coordinates":[-10,10],"type":"Point" },
	  "path":[]
	} ]
} ]
```
