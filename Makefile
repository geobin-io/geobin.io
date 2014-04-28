run:
	go run geobin.go config.go handlers.go geobinrequest.go util.go
build:
	go build geobin.go config.go handlers.go geobinrequest.go util.go
clean:
	go clean
debug:
	go build -o debug.out && ./debug.out -debug=true
