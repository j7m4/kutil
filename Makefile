.PHONY: build clean test lint vet fmt deps kctx kflap install

# Build all binaries
build: kctx kflap

# Build kctx binary
kctx:
	go build -o bin/kctx ./cmd/kctx

# Build kflap binary
kflap:
	go build -o bin/kflap ./cmd/kflap

# Clean build artifacts
clean:
	rm -rf bin/

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install binaries to system
install: kctx kflap
	sudo cp bin/kctx /usr/local/bin/kctx
	sudo chmod +x /usr/local/bin/kctx
	sudo cp bin/kflap /usr/local/bin/kflap
	sudo chmod +x /usr/local/bin/kflap

# Run all checks
check: fmt vet test

# Development setup
setup: deps install-tools

# Default target
all: clean fmt vet test build