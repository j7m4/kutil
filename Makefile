.PHONY: build clean test lint vet fmt deps kctx install

# Build all binaries
build: kctx

# Build kctx binary
kctx:
	go build -o bin/kctx ./cmd/kctx

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

# Install kctx binary to system
install: kctx
	sudo cp bin/kctx /usr/local/bin/kctx
	sudo chmod +x /usr/local/bin/kctx

# Run all checks
check: fmt vet test

# Development setup
setup: deps install-tools

# Default target
all: clean fmt vet test build