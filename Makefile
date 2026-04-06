.PHONY: build run test clean install lint fmt help

# Binary name
BINARY_NAME=cdnd
VERSION?=dev
GIT_COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFLAGS=-v

# Build flags
LDFLAGS=-ldflags "-X github.com/zwh8800/cdnd/interface/cmd.Version=$(VERSION) \
	-X github.com/zwh8800/cdnd/interface/cmd.GitCommit=$(GIT_COMMIT) \
	-X github.com/zwh8800/cdnd/interface/cmd.BuildDate=$(BUILD_DATE)"

# Default target
all: clean build

## build: Build the binary
build:
	$(GOBUILD) $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY_NAME) .

## run: Run the application
run:
	$(GOCMD) run $(LDFLAGS) . 

## test: Run tests
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-cover: Run tests with coverage report
test-cover: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## clean: Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

## install: Install the binary to $GOPATH/bin
install:
	$(GOCMD) install $(LDFLAGS) .

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOCMD) fmt ./...

## tidy: Tidy go.mod
tidy:
	$(GOMOD) tidy

## deps: Download dependencies
deps:
	$(GOMOD) download

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'

# Development targets

## dev: Build and run in development mode
dev: build
	./bin/$(BINARY_NAME)

## watch: Watch for changes and rebuild (requires entr)
watch:
	@which entr > /dev/null || (echo "Install entr: brew install entr" && exit 1)
	find . -name "*.go" -not -path "./vendor/*" | entr -r make dev

# Cross-platform builds

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows

## build-linux: Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .

## build-darwin: Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .

## build-windows: Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
