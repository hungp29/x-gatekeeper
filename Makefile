# x-gatekeeper Go service
BINARY_NAME ?= xgatekeeper
MAIN_PKG    := ./cmd/xgatekeeper
GO          := go
GOFLAGS     :=
LDFLAGS     := -s -w

.PHONY: all build build-linux test run clean deps lint

all: deps build

# Install/update dependencies (creates go.sum)
deps:
	$(GO) mod tidy
	$(GO) mod download

# Build binary (current OS)
build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) $(MAIN_PKG)

# Build for Linux (e.g. for containers)
build-linux:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)

# Run tests
test:
	$(GO) test $(GOFLAGS) ./...

# Run tests with coverage
test-coverage:
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# Run the server locally (HTTP_PORT=8080 by default)
run: build
	./bin/$(BINARY_NAME)

# Lint (requires golangci-lint if installed)
lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || $(GO) vet ./...

# Remove build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
