.PHONY: all test lint build bench fmt

all: lint test build

# Run linting
lint:
	golangci-lint run

# Run all tests
test:
	go test ./...

# Build the project
build:
	go build ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem -run=^$$ ./...

# Format code
fmt:
	go fmt ./...
