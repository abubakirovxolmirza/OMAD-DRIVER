.PHONY: help build run test clean docker-build docker-run docker-stop docker-logs migrate-up migrate-down swagger seed-db deps install-swagger fmt lint

# Color output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## Development
build: ## Build the application binary
	@echo "$(YELLOW)Building application...$(NC)"
	go build -o taxi-service cmd/main.go
	@echo "$(GREEN)✓ Build complete: ./taxi-service$(NC)"

run: ## Run the application locally
	@echo "$(YELLOW)Starting application...$(NC)"
	go run cmd/main.go

dev: ## Run in development mode with live reload (requires entr or similar)
	@echo "$(YELLOW)Starting in development mode...$(NC)"
	@go run cmd/main.go

test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and display coverage
	go tool cover -html=coverage.out

clean: ## Clean build files and uploads
	@echo "$(YELLOW)Cleaning...$(NC)"
	rm -f taxi-service
	rm -rf uploads/*
	rm -rf logs/*
	@echo "$(GREEN)✓ Clean complete$(NC)"

## Database
seed-db: ## Seed database with regions, districts, and pricing
	@echo "$(YELLOW)Seeding database...$(NC)"
	go run cmd/tools/dbseed/main.go -action=seed
	@echo "$(GREEN)✓ Database seeded$(NC)"

seed-db-force: ## Force seed database (skips confirmation)
	@echo "$(YELLOW)Force seeding database...$(NC)"
	go run cmd/tools/dbseed/main.go -action=seed -force

cleanup-db: ## Clean database (remove all data except schema)
	@echo "$(RED)WARNING: This will delete all data from the database$(NC)"
	go run cmd/tools/dbseed/main.go -action=cleanup

## Docker
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	docker build -t taxi-service:latest .
	@echo "$(GREEN)✓ Docker image built$(NC)"

docker-up: ## Start Docker containers
	@echo "$(YELLOW)Starting Docker containers...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Containers started$(NC)"
	@echo "$(YELLOW)API URL: http://localhost:8080$(NC)"

docker-down: ## Stop Docker containers
	@echo "$(YELLOW)Stopping Docker containers...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Containers stopped$(NC)"

docker-logs: ## Show Docker logs
	docker-compose logs -f app

docker-logs-db: ## Show database logs
	docker-compose logs -f db

docker-ps: ## Show running containers
	docker-compose ps

docker-rebuild: docker-down docker-build docker-up ## Rebuild and restart containers
	@echo "$(GREEN)✓ Docker containers rebuilt and restarted$(NC)"

## Dependencies
deps: ## Download and tidy dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

install-swagger: ## Install Swagger CLI
	@echo "$(YELLOW)Installing Swagger...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)✓ Swagger installed$(NC)"

## Code Quality
fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

lint: ## Run linter (requires golangci-lint)
	@echo "$(YELLOW)Running linter...$(NC)"
	golangci-lint run ./...

vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	go vet ./...

## Documentation
swagger: ## Generate Swagger documentation
	@echo "$(YELLOW)Generating Swagger docs...$(NC)"
	swag init -g cmd/main.go -o ./docs
	@echo "$(GREEN)✓ Swagger documentation generated$(NC)"
	@echo "$(YELLOW)Access Swagger UI at: http://localhost:8080/swagger/index.html$(NC)"

docs: swagger ## Alias for swagger

## Deployment
prod-build: fmt vet test build ## Full production build (format, vet, test, build)
	@echo "$(GREEN)✓ Production build complete$(NC)"

release: ## Create a production release
	@echo "$(YELLOW)Creating production release...$(NC)"
	go build -ldflags="-s -w" -o taxi-service cmd/main.go
	@echo "$(GREEN)✓ Release build complete$(NC)"

## Development Database
db-init: ## Initialize local database (requires PostgreSQL)
	@echo "$(YELLOW)Initializing database...$(NC)"
	createdb taxi_service || true
	@echo "$(GREEN)✓ Database initialized$(NC)"

db-drop: ## Drop local database
	@echo "$(RED)WARNING: This will drop the local database$(NC)"
	dropdb taxi_service || true
	@echo "$(GREEN)✓ Database dropped$(NC)"

## Health & Status
health: ## Check API health
	@curl -s http://localhost:8080/health | jq . || echo "API is not running"

status: docker-ps ## Show application status

version: ## Show version info
	@go version

## Helpers
install: deps build ## Install dependencies and build application
	@echo "$(GREEN)✓ Installation complete$(NC)"

.PHONY: all
all: clean deps fmt vet test build ## Run full development cycle
	@echo "$(GREEN)✓ Full cycle complete$(NC)"
