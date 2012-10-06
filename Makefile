# Expects that this directory is in your GOPATH

# Executables
build:
	[ -d bin ] || mkdir -p bin
	go build -o bin/ebooker_server ebooker/server
	go build -o bin/ebooker_client ebooker/client


# Double dollar sign on sed command since "$$" -> "$" in a Makefile
fmt:
	go fmt ebooker/client
	go fmt ebooker/server
	go fmt ebooker/defs
	go fmt ebooker/logging
	go fmt ebooker/oauth1
	find src -name "*.go" -exec sed -i 's/[ \t]*$$//' \{\} \;


test:
	go test ebooker/server
	go test ebooker/oauth1
	find src -name "*.db" -exec rm \{\} \;

clean:
	rm -rf bin/*
