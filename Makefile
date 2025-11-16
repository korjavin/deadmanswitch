.PHONY: help lint lint-fix lint-install test build run clean

# Default target
help:
	@echo "Available targets:"
	@echo "  lint          - Run golangci-lint"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  lint-install  - Install golangci-lint"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  clean         - Clean build artifacts"

# Linting targets
lint:
	golangci-lint run --timeout=5m

lint-fix:
	golangci-lint run --fix --timeout=5m

lint-install:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.1.2
	@echo "golangci-lint installed to $$(go env GOPATH)/bin/golangci-lint"
	@echo "Make sure $$(go env GOPATH)/bin is in your PATH"

# Testing targets
test:
	go test ./...

test-coverage:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt

test-coverage-html:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Build targets
build:
	go build -o bin/deadmanswitch ./cmd/server

run:
	go run ./cmd/server

# Clean targets
clean:
	rm -rf bin/
	rm -f coverage.txt coverage.html
	rm -f *.db
