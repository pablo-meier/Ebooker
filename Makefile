# Expects that this directory is in your GOPATH

# Executables
build:
	[ -d bin ] || mkdir -p bin
	go build -o bin/ebooker_server ebooker/server
	go build -o bin/ebooker_client ebooker/client

fmt:
	go fmt ebooker/client
	go fmt ebooker/server
	go fmt ebooker/defs
	go fmt ebooker/logging
	go fmt ebooker/oauth1

test:
	go test ebooker/server
	go test ebooker/oauth

clean:
	rm -rf bin/*
