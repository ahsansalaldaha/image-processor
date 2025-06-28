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

### Running the System

1. Start the infrastructure services:
```bash
docker-compose up rabbitmq minio postgres -d
```

2. Build and run all services:
```bash
docker-compose up --build
```

### Individual Service Development

Each service can be developed and tested independently:

```bash
# Run only url-ingestor
docker-compose up url-ingestor

# Run only image-fetcher
docker-compose up image-fetcher

# Run only image-metadata
docker-compose up image-metadata
```

### Environment Variables

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