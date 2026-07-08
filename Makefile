.PHONY: build test lint security-lint vulncheck verify-tidy check install clean deps help

# Binary name
BINARY_NAME=tpm
EXAMPLE_VALIDATOR_DIR=examples/spring-boot-health-checks/validator
GOVULNCHECK=go run golang.org/x/vuln/cmd/govulncheck@latest

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

# Run security lint through golangci-lint/gosec.
security-lint:
	@echo "Running security linter..."
	@if ! command -v golangci-lint > /dev/null; then \
		echo "golangci-lint is required for security-lint" >&2; \
		exit 1; \
	fi
	golangci-lint run --enable-only=gosec --max-same-issues=0 ./...

# Run Go vulnerability checks for every Go module in this repo.
vulncheck:
	@echo "Running vulnerability checks..."
	$(GOVULNCHECK) ./...
	cd $(EXAMPLE_VALIDATOR_DIR) && $(GOVULNCHECK) ./...

# Verify all Go module files are tidy.
verify-tidy:
	@echo "Verifying go.mod files are tidy..."
	go mod tidy -diff
	cd $(EXAMPLE_VALIDATOR_DIR) && go mod tidy -diff

# Run the full local quality gate before opening or updating a PR.
check: verify-tidy lint security-lint test vulncheck build

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
	cd $(EXAMPLE_VALIDATOR_DIR) && go mod download && go mod tidy

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the $(BINARY_NAME) binary"
	@echo "  test          - Run tests"
	@echo "  lint          - Run linter"
	@echo "  security-lint - Run gosec through golangci-lint"
	@echo "  vulncheck     - Run govulncheck for repo Go modules"
	@echo "  verify-tidy   - Verify go.mod and go.sum are tidy"
	@echo "  check         - Run the full local PR quality gate"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  help          - Display this help message"
