build: server client
server:
	mkdir -p bin
	go build -o 'bin/server' ./cmds/server

client:
	mkdir -p bin
	go build -o 'bin/client' ./cmds/client

dep:
	godep save ./...

.PHONY: build server client
