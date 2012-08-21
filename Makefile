

build:
	[ -d bin ] || mkdir -p bin
	go build -o bin/ebooker src/markov_consumer.go

test:
	go test src/markov_consumer_test.go

clean:
	rm -rf bin/*
