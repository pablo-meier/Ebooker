# Expects that this directory is in your GOPATH

build:
	[ -d bin ] || mkdir -p bin
	go build -o bin/ebooker_server ebooker/server
	go build -o bin/ebooker_client ebooker/client

fmt:
	go fmt ebooker/ebooks
	go fmt ebooker/client
	go fmt ebooker/server

test:
	go test ebooker/ebooks

clean:
	rm -rf bin/*
