.PHONY: help build run test clean docker-build docker-run migrate-up migrate-down swagger

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	go build -o taxi-service cmd/main.go

run: ## Run the application
	go run cmd/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build files
	rm -f taxi-service
	rm -rf uploads/*

swagger: ## Generate swagger documentation
	swag init -g cmd/main.go -o ./docs

migrate-up: ## Run database migrations up
	go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	go run cmd/migrate/main.go down

docker-build: ## Build docker image
	docker build -t taxi-service:latest .

docker-run: ## Run docker container
	docker-compose up -d

deps: ## Download dependencies
	go mod download
	go mod tidy

install-swagger: ## Install swagger CLI
	go install github.com/swaggo/swag/cmd/swag@latest
