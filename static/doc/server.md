# Geobin Server

The Geobin server hosts the Geobin [web client] as well as the [API]. It is written in [go] and uses [redis] so, assuming you have a working go [dev environment] and [redis server] running the following should get you up and running.

## Setup

```bash
> go get github.com/geoloqi/geobin-go
> cd $GOPATH/src/github.com/geoloqi/geobin-go
> make setup # (runs `go get -t` and `npm install`)
> cp config.json.dist config.json
```

## Configure

Geobin comes with a default `config.json` file which should get you up and running for development, you'll just need to copy `config.json.dist` to `config.json` (as listed in the Setup steps above).

See the comments in `config.json.dist` for documentation of the available keys.

## Run

```bash
> make run
```

## Test

```bash
> go test
```

## Build

```bash
> go build -o geobin
```

[go]: http://golang.org
[dev environment]: http://golang.org/doc/install
[redis]: http://redis.io
[redis server]: http://redis.io/download
[web client]: client.md
[API]: api.md
