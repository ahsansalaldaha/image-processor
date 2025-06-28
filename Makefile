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