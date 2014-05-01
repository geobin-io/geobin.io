geobin-go
=========

## setup

1. `go get github.com/geoloqi/geobin-go`
1. `cd $GOPATH/src/github.com/geoloqi/geobin-go`
1. `make setup` (runs `go get -t` and `npm install`)

## run

1. `make run`
1. `open http://localhost:8080`
1. create a bin
1. `curl -i -X POST -H "Content-Type: application/json" -d @gtCallback.json http://localhost:8080/BIN_ID`
1. check you bin

## test

### server

1. `go test`

### client

See [client readme](static/app/README.md)

## build

1. `go build -o geobin`
