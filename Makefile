build: server client
server:
	mkdir -p bin8080
	go build -o 'bin8080/server' ./cmds/server

client:
	mkdir -p bin8080
	go build -o 'bin8080/client' ./cmds/client

dep:
	godep save ./...

.PHONY: build server client
