.PHONY: fmt test vet check build clean lint coverage install help

BINARY := gdashlint

help:
	@echo "Available targets:"
	@echo "  fmt       - Format code with gofmt"
	@echo "  lint      - Run golangci-lint checks"
	@echo "  vet       - Run go vet"
	@echo "  test      - Run tests"
	@echo "  coverage  - Generate and display test coverage"
	@echo "  check     - Run all checks (fmt, vet, test)"
	@echo "  build     - Build the binary"
	@echo "  install   - Install the binary to \$$GOBIN"
	@echo "  clean     - Remove build artifacts"

fmt:
	gofmt -w ./cmd ./internal

lint:
	golangci-lint run ./...

vet:
	go vet ./...

test:
	go test -v ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

check: fmt vet test

build:
	go build -o bin/$(BINARY) ./cmd/gdashlint

install:
	go install ./cmd/gdashlint

clean:
	rm -rf bin coverage.out coverage.html
