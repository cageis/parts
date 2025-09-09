.DEFAULT_GOAL := help

# Build variables
GO = go
BINARY_NAME = parts
BUILD_DIR = bin
SRC_DIR = src

# Source files
SOURCES = $(wildcard $(SRC_DIR)/*.go)

# Build flags
LDFLAGS = -w -s
BUILD_FLAGS = -ldflags="$(LDFLAGS)"

.PHONY: help build test test-coverage test-coverage-html benchmark clean install quickstart ssh ssh-dry fmt vet check examples

help: ## Show this help message
	@echo "Parts - SSH Config Partials Manager"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m  %-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build                    # Build the binary"
	@echo "  make test                     # Run all tests"
	@echo "  make quickstart               # Build and run SSH config merge"
	@echo "  make ssh                      # Merge SSH config partials"
	@echo "  make ssh-dry                  # Preview SSH config changes (dry-run)"
	@echo "  make examples                 # Run all usage examples"

build: $(BUILD_DIR)/$(BINARY_NAME) ## Build the parts binary

$(BUILD_DIR)/$(BINARY_NAME): $(SOURCES) main.go | $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) -o $@ ./main.go

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

test: ## Run all tests
	$(GO) test -v ./src

test-coverage: ## Run tests with coverage report
	$(GO) test -v -cover ./src

test-coverage-html: ## Generate HTML coverage report
	$(GO) test -v -coverprofile=coverage.out ./src && $(GO) tool cover -html=coverage.out -o coverage.html

benchmark: ## Run benchmarks
	$(GO) test -bench=. ./src

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	$(GO) clean

install: build ## Install parts to GOPATH/bin
	$(GO) install ./main.go

quickstart: build ssh ## Build and run SSH config merge

ssh: $(BUILD_DIR)/$(BINARY_NAME) ## Merge SSH config partials using default paths
	./$(BUILD_DIR)/$(BINARY_NAME) ~/.ssh/config ~/.ssh/config.d "#"

ssh-dry: $(BUILD_DIR)/$(BINARY_NAME) ## Preview SSH config changes (dry-run mode)
	./$(BUILD_DIR)/$(BINARY_NAME) --dry-run ~/.ssh/config ~/.ssh/config.d "#"

fmt: ## Format Go code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

check: fmt vet test ## Run all checks (format, vet, test)

examples: build ## Run all usage examples
	cd examples && ./run-all-examples.sh