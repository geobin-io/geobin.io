# Geobin Server

The Geobin server hosts the Geobin [web client] as well as the [API]. It is written in [go] so, assuming you have a working go [dev environment] the following should get you up and running.

## Setup

```bash
> go get github.com/geoloqi/geobin-go
> cd $GOPATH/src/github.com/geoloqi/geobin-go
> make setup # (runs `go get -t` and `npm install`)
```

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
[web client]: client.md
[API]: api.md
