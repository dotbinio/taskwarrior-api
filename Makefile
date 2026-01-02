.PHONY: build run test clean dev install swagger docker-build docker-push

# Build the server binary
build: swagger
	go build -o bin/taskwarrior-api ./cmd/server

# Run the server
run: swagger
	go run ./cmd/server

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Generate Swagger documentation
swagger:
	swag init -g cmd/server/main.go -o docs

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Install dependencies
install:
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Docker commands
docker-build:
	docker build -t taskwarrior-api:latest .

docker-push:
	@echo "Tag and push to your registry:"
	@echo "  docker tag taskwarrior-api:latest your-registry/taskwarrior-api:latest"
	@echo "  docker push your-registry/taskwarrior-api:latest"

# Kubernetes deployment
k8s-deploy:
	kubectl apply -f k8s/deployment.yaml

docker-down:
	docker-compose down

k8s-logs:
	kubectl logs -f -n taskwarrior deployment/taskwarrior-api

docker: docker-build docker-up

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  run            - Run the server"
	@echo "  dev            - Run with hot reload (requires air)"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  install        - Install/update dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start Docker container"
	@echo "  docker-down    - Stop Docker container"
	@echo "  docker-logs    - View Docker logs"
	@echo "  docker         - Build and start Docker container"

