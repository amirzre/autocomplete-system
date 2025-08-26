# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
APP_NAME = autocomplete-system

.PHONY: help build run test docker-build docker-dev-up docker-dev-stop docker-prod-up docker-prod-stop

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

## Docker
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME) .

.PHONY: docker-dev-up
docker-dev-up: ## Run development environment with docker compose
	@echo "Starting development environment..."
	docker compose -f docker-compose.dev.yml up -d

.PHONY: docker-dev-stop
docker-dev-stop: ## Stop Docker dev containers
	@echo "Stopping Docker containers..."
	docker compose -f docker-compose.dev.yml down

.PHONY: docker-prod-up
docker-prod-up: ## Run production environment with docker compose
	@echo "Starting production environment..."
	docker compose up --build -d

.PHONY: docker-prod-stop
docker-prod-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker compose down
