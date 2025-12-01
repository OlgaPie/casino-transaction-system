.PHONY: help build test coverage test-integration run-docker stop clean deps

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## Build API and Consumer binaries
	@echo "Building API server..."
	@go build -o api_server ./cmd/api
	@echo "Building Consumer app..."
	@go build -o consumer_app ./cmd/consumer
	@echo "✓ Build complete"

test: ## Run all tests
	go test -v ./...

coverage: ## Generate test coverage report and open in browser
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-integration: ## Run integration tests (requires Docker)
	go test -v -run Integration ./...

run-docker: ## Start all services with Docker Compose
	docker-compose up --build

stop: ## Stop Docker Compose services
	docker-compose down

clean: ## Remove build artifacts, coverage files, and Docker volumes
	@echo "Cleaning build artifacts..."
	@rm -f api_server consumer_app coverage.out coverage_new.out
	@echo "Stopping and removing Docker containers..."
	@docker-compose down -v
	@echo "✓ Cleanup complete"

deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

.DEFAULT_GOAL := help
