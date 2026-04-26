.PHONY: help dev build test clean migrate-up migrate-down docker-build docker-push deploy

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start development environment with docker-compose
	docker-compose up -d
	@echo "Development environment started. Access at http://localhost:8080"
	@echo "Database running at localhost:5432"

dev-stop: ## Stop development environment
	docker-compose down

dev-reload: ## Rebuild and restart development environment
	docker-compose down
	docker-compose up -d --build

build: ## Build the application binary
	go build -o bin/server ./cmd/server

build-linux: ## Build Linux binary locally
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server-linux ./cmd/server

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests with Docker
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

migrate-up: ## Run database migrations
	docker-compose exec db goose -dir=/docker-entrypoint-initdb.d up

migrate-down: ## Rollback database migrations
	docker-compose exec db goose -dir=/docker-entrypoint-initdb.d down

migrate-new: ## Create a new migration file
	@read -p "Enter migration name: " name; \
	docker-compose exec db goose -dir=/docker-entrypoint-initdb.d create $$name sql

docker-build: ## Build Docker image
	docker build -t todo-app:latest .

docker-push: ## Push Docker image to registry (set IMAGE_TAG)
	docker tag todo-app:latest ${IMAGE_TAG}
	docker push ${IMAGE_TAG}

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...
	go vet ./...
	gosimports -w .
	gofumpt -w .

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/segmentio/golines@latest
	go install github.com/daixiang0/gci@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/rinchsan/gosimports/cmd/gosimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed successfully!"

airdeps: ## Download Air TOML config if missing
	@if [ ! -f .air.toml ]; then \
		curl -sL https://raw.githubusercontent.com/cosmtrek/air/master/.air.toml.example -o .air.toml; \
		echo "[Custom air configuration created]"; \
	fi

setup: airdeps install-tools ## Setup development environment
	cp .env.example .env
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start the application"

logs: ## Show application logs
	docker-compose logs -f app

db-logs: ## Show database logs
	docker-compose logs -f db

db-shell: ## Connect to database
	docker-compose exec db psql -U todo_user -d todo_db

swagger: ## Generate Swagger documentation (if using swag)
	swag init -g cmd/server/main.go -o docs/swagger

proto: ## Generate protobuf files (if using protobuf)
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/*.proto

.PHONY: all
all: fmt lint test build ## Run all checks and build