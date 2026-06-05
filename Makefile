.PHONY: build test lint install clean help

# Binary name
BINARY_NAME=tpm

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping..."; \
		go vet ./...; \
	fi

# Install binary to GOPATH
install:
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Display help
help:
	@echo "Available targets:"
	@echo "  build   - Build the $(BINARY_NAME) binary"
	@echo "  test    - Run tests"
	@echo "  lint    - Run linter"
	@echo "  install - Install binary to GOPATH/bin"
	@echo "  clean   - Remove build artifacts"
	@echo "  deps    - Download and tidy dependencies"
	@echo "  help    - Display this help message"
