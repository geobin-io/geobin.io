run:
	go run geobin.go config.go handlers.go geobinrequest.go util.go
debug:
	go build -o debug.out && ./debug.out -debug=true
