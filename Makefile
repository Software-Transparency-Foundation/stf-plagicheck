# Makefile for plagicheck

# Variables
BINARY_NAME=plagicheck
VERSION?=1.0.0
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO=go
GOFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT)"
SRC_DIR=./cmd
PKG_DIR=./pkg
TEST_DIR=./test

# Targets
.PHONY: all build test clean install help

all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME) $(VERSION) (commit: $(GIT_COMMIT))..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(SRC_DIR)/main.go
	@echo "Build complete: $(BINARY_NAME)"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GO) test -v $(PKG_DIR)/...
	@echo "Tests complete"

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out $(PKG_DIR)/...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## install: Install binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "Install complete"

## run-tests: Run integration tests using test files
run-tests: build
	@echo "Running integration tests..."
	@echo "\n=== Test 1: File match ==="
	./$(BINARY_NAME) $(TEST_DIR)/file_match.wfp
	@echo "\n=== Test 2: Snippet match ==="
	./$(BINARY_NAME) $(TEST_DIR)/snippet_match.wfp
	@echo "\n=== Test 3: Generate WFP from file ==="
	./$(BINARY_NAME) -fp $(TEST_DIR)/test-file_snippet.cpp | head -5
	@echo "\n=== Test 4: Scan with min-hits threshold ==="
	./$(BINARY_NAME) --min-hits 5 $(TEST_DIR)/snippet_match.wfp
	@echo "\n=== Test 5: Version ==="
	./$(BINARY_NAME) --version
	@echo "\nIntegration tests complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...
	@echo "Vet complete"

## check: Run fmt, vet and test
check: fmt vet test

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
