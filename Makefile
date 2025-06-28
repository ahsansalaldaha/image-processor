.PHONY: help dev prod build clean logs restart stop

# Default target
help:
	@echo "Available commands:"
	@echo "  dev     - Start development environment with hot reloading"
	@echo "  prod    - Start production environment"
	@echo "  build   - Build all services"
	@echo "  logs    - Show logs for all services"
	@echo "  restart - Restart all services"
	@echo "  stop    - Stop all services"
	@echo "  clean   - Clean up containers, volumes, and images"
	@echo ""
	@echo "Testing:"
	@echo "  test                 - Run unit tests"
	@echo "  test-coverage        - Run tests with coverage report"
	@echo "  test-integration     - Run integration tests"
	@echo "  test-postgres-upgrade - Test PostgreSQL 17 upgrade"
	@echo ""
	@echo "Monitoring:"
	@echo "  health-check         - Check service health"
	@echo "  status-check         - Check service status"
	@echo "  metrics-check        - Check metrics endpoints"
	@echo "  open-monitoring      - Open monitoring UIs"

# Development environment
dev:
	@echo "Starting development environment with hot reloading..."
	docker-compose -f docker-compose.dev.yml up --build

# Production environment
prod:
	@echo "Starting production environment..."
	docker-compose -f docker-compose.prod.yml up --build -d

# Build all services
build:
	@echo "Building all services..."
	docker-compose -f docker-compose.dev.yml build

# Show logs
logs:
	@echo "Showing logs for all services..."
	docker-compose -f docker-compose.dev.yml logs -f

# Restart services
restart:
	@echo "Restarting all services..."
	docker-compose -f docker-compose.dev.yml restart

# Stop services
stop:
	@echo "Stopping all services..."
	docker-compose -f docker-compose.dev.yml down

# Clean up everything
clean:
	@echo "Cleaning up containers, volumes, and images..."
	docker-compose -f docker-compose.dev.yml down -v --rmi all
	docker-compose -f docker-compose.prod.yml down -v --rmi all
	docker system prune -f

# Development with specific service
dev-url-ingestor:
	@echo "Starting url-ingestor service in development mode..."
	docker-compose -f docker-compose.dev.yml up --build url-ingestor

dev-image-fetcher:
	@echo "Starting image-fetcher service in development mode..."
	docker-compose -f docker-compose.dev.yml up --build image-fetcher

dev-image-metadata:
	@echo "Starting image-metadata service in development mode..."
	docker-compose -f docker-compose.dev.yml up --build image-metadata

# Production with specific service
prod-url-ingestor:
	@echo "Starting url-ingestor service in production mode..."
	docker-compose -f docker-compose.prod.yml up --build -d url-ingestor

prod-image-fetcher:
	@echo "Starting image-fetcher service in production mode..."
	docker-compose -f docker-compose.prod.yml up --build -d image-fetcher

prod-image-metadata:
	@echo "Starting image-metadata service in production mode..."
	docker-compose -f docker-compose.prod.yml up --build -d image-metadata

# Testing commands
test:
	go test ./... -v

test-coverage:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	./test/integration_test.sh

test-postgres-upgrade:
	@echo "Testing PostgreSQL 17 upgrade..."
	./test/postgres_upgrade_test.sh

# Monitoring commands
monitor-logs:
	docker-compose -f docker-compose.dev.yml logs -f

monitor-url-ingestor:
	docker-compose -f docker-compose.dev.yml logs -f url-ingestor

monitor-image-fetcher:
	docker-compose -f docker-compose.dev.yml logs -f image-fetcher

monitor-image-metadata:
	docker-compose -f docker-compose.dev.yml logs -f image-metadata

# Health checks
health-check:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq .
	@curl -s http://localhost:8081/health | jq .
	@curl -s http://localhost:8082/health | jq .

status-check:
	@echo "Checking service status..."
	@curl -s http://localhost:8080/status | jq .
	@curl -s http://localhost:8080/queue/status | jq .

metrics-check:
	@echo "Checking metrics endpoints..."
	@curl -s http://localhost:8080/metrics | head -20
	@echo "---"
	@curl -s http://localhost:8081/metrics | head -20
	@echo "---"
	@curl -s http://localhost:8082/metrics | head -20

# Open monitoring UIs
open-monitoring:
	@echo "Opening monitoring UIs..."
	@open http://localhost:3000  # Grafana
	@open http://localhost:9090  # Prometheus
	@open http://localhost:16686 # Jaeger
	@open http://localhost:15672 # RabbitMQ
	@open http://localhost:9001  # MinIO 