# PTD - Portable Tournament Data Makefile
# Suparena Software Inc.

# Variables
BINARY_NAME=ptd
GO=go
GOTEST=$(GO) test
GOCOVER=$(GO) tool cover
GOFMT=gofmt
GOLINT=golangci-lint
GOMOD=$(GO) mod
BUILD_DIR=build
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

.PHONY: all build test clean help fmt lint coverage bench install deps tidy check security

## help: Display this help message
help:
	@echo "PTD Makefile Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make ${GREEN}<target>${NC}\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  ${YELLOW}%-15s${NC} %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## all: Build and test everything
all: deps fmt lint test build

## build: Build the PTD library
build:
	@echo "$(GREEN)Building PTD library...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build -v ./...
	@echo "$(GREEN)Build complete!$(NC)"

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	@echo "$(GREEN)Tests complete!$(NC)"

## test-short: Run short tests only
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) -v -short ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "$(GREEN)Running tests with verbose output...$(NC)"
	$(GOTEST) -v -race -cover ./...

## coverage: Generate test coverage report
coverage: test
	@echo "$(GREEN)Generating coverage report...$(NC)"
	$(GOCOVER) -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	$(GOCOVER) -func=$(COVERAGE_FILE)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"
	@echo "Opening coverage report..."
	@open $(COVERAGE_HTML) 2>/dev/null || xdg-open $(COVERAGE_HTML) 2>/dev/null || echo "Please open $(COVERAGE_HTML) manually"

## bench: Run benchmarks
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) -w .
	$(GO) fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@which $(GOLINT) > /dev/null || (echo "$(RED)golangci-lint not installed. Run: make install-lint$(NC)" && exit 1)
	$(GOLINT) run ./...
	@echo "$(GREEN)Linting complete!$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)Vet complete!$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	@echo "$(GREEN)Dependencies downloaded!$(NC)"

## tidy: Tidy go.mod and go.sum
tidy:
	@echo "$(GREEN)Tidying go.mod and go.sum...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Tidy complete!$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f *.ptd *.test
	@echo "$(GREEN)Clean complete!$(NC)"

## install: Install the library
install:
	@echo "$(GREEN)Installing PTD library...$(NC)"
	$(GO) install ./...
	@echo "$(GREEN)Installation complete!$(NC)"

## install-lint: Install golangci-lint
install-lint:
	@echo "$(GREEN)Installing golangci-lint...$(NC)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "$(GREEN)golangci-lint installed!$(NC)"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed!$(NC)"

## security: Run security scan
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@which gosec > /dev/null || (echo "$(YELLOW)Installing gosec...$(NC)" && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	gosec -fmt json -out security-report.json ./... || true
	@echo "$(GREEN)Security scan complete! Check security-report.json$(NC)"

## docs: Generate documentation
docs:
	@echo "$(GREEN)Generating documentation...$(NC)"
	@which godoc > /dev/null || (echo "$(YELLOW)Installing godoc...$(NC)" && go install golang.org/x/tools/cmd/godoc@latest)
	@echo "$(GREEN)Documentation server starting at http://localhost:6060$(NC)"
	@echo "$(YELLOW)Press Ctrl+C to stop$(NC)"
	godoc -http=:6060

## example: Run example code
example:
	@echo "$(GREEN)Running example...$(NC)"
	$(GO) test -v -run Example

## ci: Run CI pipeline locally
ci: deps check coverage
	@echo "$(GREEN)CI pipeline complete!$(NC)"

## version: Display version information
version:
	@echo "PTD Version Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Commit:  $(COMMIT)"
	@echo "  Built:   $(BUILD_TIME)"

# Default target
.DEFAULT_GOAL := help