setup:
	go get -t && npm install
run:
	go run geobin.go config.go handlers.go geobinrequest.go util.go socket.go socketmap.go
debug:
	go build -o debug.out && ./debug.out -debug=true
