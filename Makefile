# Expects that this directory is in your GOPATH

build:
	[ -d bin ] || mkdir -p bin
	go build markov

test:
	go test markov

clean:
	rm -rf bin/*
