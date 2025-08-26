# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
APP_NAME = autocomplete-system

.PHONY: help build run test

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'


# Build the application
build: ## Build the application
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -o bin/$(APP_NAME) ./cmd/server
	@echo "Build complete: bin/$(APP_NAME)"


# Run the application locally
run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	$(GOCMD) run ./cmd/server


# Run tests
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race ./...
