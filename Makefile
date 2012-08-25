# Expects that this directory is in your GOPATH

build:
	[ -d bin ] || mkdir -p bin
	go build -o bin/ebooker 

test:
	go test markov

clean:
	rm -rf bin/*
