# Image Processing System

A microservices-based image processing system built with Go. The system consists of three main services that work together to process images from URLs.

## Architecture

### Services

1. **url-ingestor** - HTTP API service that accepts image URLs and queues them for processing
2. **image-fetcher** - Worker service that downloads, processes (grayscale), and uploads images to MinIO
3. **image-metadata** - Service that stores metadata about processed images in PostgreSQL

### Infrastructure

- **RabbitMQ** - Message queue for inter-service communication
- **MinIO** - Object storage for processed images
- **PostgreSQL** - Database for storing image metadata

## Service-Specific Configuration

Each service has its own configuration that only includes what it needs:

- **url-ingestor**: Server port, RabbitMQ URL
- **image-fetcher**: RabbitMQ URL, MinIO config, Database config
- **image-metadata**: RabbitMQ URL, Database config

## Development

### Prerequisites

- Docker and Docker Compose
- Go 1.21+

### Quick Start

#### Development Mode (with Hot Reloading)

For seamless development with automatic code reloading:

```bash
# Start all services in development mode
make dev

# Or use docker-compose directly
docker-compose -f docker-compose.dev.yml up --build
```

**Features:**
- Hot reloading using Air - code changes automatically rebuild and restart services
- Volume mounts for live code editing
- Development-specific configurations
- Health checks for dependencies

#### Production Mode

For production deployment:

```bash
# Start all services in production mode
make prod

# Or use docker-compose directly
docker-compose -f docker-compose.prod.yml up --build -d
```

**Features:**
- Optimized builds without development dependencies
- Multiple replicas for scalability
- Restart policies
- Environment variable support

### Individual Service Development

Each service can be developed and tested independently:

```bash
# Development mode - individual services
make dev-url-ingestor
make dev-image-fetcher
make dev-image-metadata

# Production mode - individual services
make prod-url-ingestor
make prod-image-fetcher
make prod-image-metadata
```

### Useful Commands

```bash
# Show all available commands
make help

# Build all services
make build

# View logs
make logs

# Restart services
make restart

# Stop all services
make stop

# Clean up everything (containers, volumes, images)
make clean
```

### Environment Variables

Copy the example environment file and customize as needed:

```bash
cp env.dev.example .env.dev
```

Each service uses only the environment variables it needs:

#### url-ingestor
- `SERVER_PORT` - HTTP server port (default: 8080)
- `RABBITMQ_URL` - RabbitMQ connection URL

#### image-fetcher
- `RABBITMQ_URL` - RabbitMQ connection URL
- `MINIO_ENDPOINT` - MinIO server endpoint
- `MINIO_ACCESS_KEY` - MinIO access key
- `MINIO_SECRET_KEY` - MinIO secret key
- `MINIO_USE_SSL` - Whether to use SSL for MinIO
- `MINIO_BUCKET` - MinIO bucket name
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_USER` - PostgreSQL username
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - PostgreSQL database name
- `DB_SSLMODE` - PostgreSQL SSL mode

#### image-metadata
- `RABBITMQ_URL` - RabbitMQ connection URL
- `DB_HOST` - PostgreSQL host
- `DB_PORT` - PostgreSQL port
- `DB_USER` - PostgreSQL username
- `DB_PASSWORD` - PostgreSQL password
- `DB_NAME` - PostgreSQL database name
- `DB_SSLMODE` - PostgreSQL SSL mode

## API Endpoints

### url-ingestor (Port 8080)

- `POST /images` - Submit image URLs for processing
  - Body: `{"urls": ["http://example.com/image1.jpg", "http://example.com/image2.jpg"]}`

## Message Flow

1. Client submits image URLs to url-ingestor
2. url-ingestor publishes messages to RabbitMQ queue "image.urls"
3. image-fetcher consumes messages, downloads images, processes them, and uploads to MinIO
4. image-fetcher publishes results to RabbitMQ queue "image.processed"
5. image-metadata consumes processed messages and stores metadata in PostgreSQL

## Development Workflow

### Hot Reloading

The development environment uses [Air](https://github.com/cosmtrek/air) for hot reloading:

- Code changes automatically trigger rebuilds
- Services restart automatically
- No manual rebuilds needed during development

### File Structure

```
├── docker-compose.dev.yml      # Development environment
├── docker-compose.prod.yml     # Production environment
├── docker/
│   ├── *.dev.Dockerfile        # Development Dockerfiles with Air
│   └── *.Dockerfile            # Production Dockerfiles
├── .air.toml                   # Air config for url-ingestor
├── air-image-fetcher.toml      # Air config for image-fetcher
├── air-image-metadata.toml     # Air config for image-metadata
└── Makefile                    # Convenient commands
```

### Production Deployment

For production deployment:

1. Set up environment variables
2. Use production docker-compose file
3. Configure external databases and message queues if needed
4. Set up proper monitoring and logging

```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up --build -d
``` 