.PHONY: fmt test vet check build clean

BINARY := gdashlint

fmt:
	gofmt -w ./cmd ./internal

test:
	go test ./...

vet:
	go vet ./...

check: fmt vet test

build:
	go build -o bin/$(BINARY) ./cmd/gdashlint

clean:
	rm -rf bin
