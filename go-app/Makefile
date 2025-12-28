# MongoBleed Go Implementation Makefile

BINARY_NAME=mongobleed
SOURCE_FILE=mongobleed.go

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) $(SOURCE_FILE)

# Build for multiple platforms
build-all: build-linux build-windows build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 $(SOURCE_FILE)
	GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 $(SOURCE_FILE)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe $(SOURCE_FILE)
	GOOS=windows GOARCH=arm64 go build -o dist/$(BINARY_NAME)-windows-arm64.exe $(SOURCE_FILE)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 $(SOURCE_FILE)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 $(SOURCE_FILE)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

# Run tests (if any)
test:
	go test -v

# Install dependencies (none required for standard library)
deps:
	go mod tidy

# Format code
fmt:
	go fmt $(SOURCE_FILE)

# Run go vet
vet:
	go vet $(SOURCE_FILE)

# Lint code (requires golint)
lint:
	golint $(SOURCE_FILE)

# Show help
help:
	@echo "MongoBleed Go Implementation"
	@echo "============================="
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build the binary for current platform"
	@echo "  build-all   - Build for all platforms (linux, windows, macOS)"
	@echo "  build-linux - Build for Linux (amd64, arm64)"
	@echo "  build-windows - Build for Windows (amd64, arm64)"
	@echo "  build-darwin - Build for macOS (amd64, arm64)"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  fmt         - Format code"
	@echo "  vet         - Run go vet"
	@echo "  lint        - Run golint"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Usage after build:"
	@echo "  ./$(BINARY_NAME) --host <target> --port 27017"
	@echo ""
	@echo "Example:"
	@echo "  make build"
	@echo "  ./$(BINARY_NAME) --host localhost --max-offset 50000"

.PHONY: all build build-all build-linux build-windows build-darwin clean test deps fmt vet lint help
