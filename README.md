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
- **PostgreSQL 17** - Database for storing image metadata

### PostgreSQL 17 Upgrade

The system has been upgraded to PostgreSQL 17, the latest stable version. This upgrade provides:

#### Benefits
- **Performance Improvements**: Enhanced query performance and better parallel processing
- **Security Enhancements**: Latest security patches and improved authentication
- **Monitoring & Observability**: Better built-in monitoring capabilities
- **Latest Features**: Access to PostgreSQL 17's new features and optimizations

#### Compatibility
- **GORM Auto-Migration**: The application uses GORM's auto-migration feature, ensuring seamless schema updates
- **Updated Dependencies**: Go dependencies have been updated to ensure full compatibility
- **No Manual Migration Required**: The upgrade is handled automatically by the application

#### Testing
A dedicated test script is available to verify the PostgreSQL 17 upgrade:

```bash
# Test PostgreSQL 17 upgrade
./test/postgres_upgrade_test.sh
```

This test verifies:
- PostgreSQL 17 is running correctly
- Database connectivity is working
- GORM auto-migration is successful
- Application startup with PostgreSQL 17

### Monitoring Stack

- **Prometheus** - Metrics collection and storage
- **Grafana** - Metrics visualization and dashboards
- **Jaeger** - Distributed tracing
- **RabbitMQ Management UI** - Queue monitoring
- **MinIO Console** - Object storage management

## Service-Specific Configuration

Each service has its own configuration that only includes what it needs:

- **url-ingestor**: Server port, RabbitMQ URL
- **image-fetcher**: RabbitMQ URL, MinIO config, Database config
- **image-metadata**: RabbitMQ URL, Database config

## Development

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- curl (for testing)
- jq (for JSON formatting)

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
- Full monitoring stack (Prometheus, Grafana, Jaeger)

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

## Usage

To submit URLs for image processing, use the following `curl` command:

```bash
curl -X POST http://localhost:8080/submit \
  -H "Content-Type: application/json" \
  -d '{
    "urls": [
      "https://demonslayer-anime.com/portal/assets/img/img_kv_2.jpg",
      "https://demonslayer-hinokami.sega.com/img/purchase/digital-standard.jpg"
    ]
  }'
```

## Monitoring & Observability

### Health Checks

All services expose health check endpoints:

```bash
# Check all service health
make health-check

# Individual service health
curl http://localhost:8080/health  # url-ingestor
curl http://localhost:8081/health  # image-fetcher
curl http://localhost:8082/health  # image-metadata
```

### Metrics

All services expose Prometheus metrics:

```bash
# Check metrics endpoints
make metrics-check

# Individual service metrics
curl http://localhost:8080/metrics  # url-ingestor
curl http://localhost:8081/metrics  # image-fetcher
curl http://localhost:8082/metrics  # image-metadata
```

### Status & Queue Monitoring

```bash
# Check service status and queue information
make status-check

# Individual endpoints
curl http://localhost:8080/status      # Service status
curl http://localhost:8080/queue/status # Queue information
curl http://localhost:8080/stats        # System statistics
```

### Monitoring UIs

Open all monitoring interfaces:

```bash
make open-monitoring
```

**Available UIs:**
- **Grafana**: http://localhost:3000 (admin/admin) - Metrics dashboards
- **Prometheus**: http://localhost:9090 - Metrics collection
- **Jaeger**: http://localhost:16686 - Distributed tracing
- **RabbitMQ Management**: http://localhost:15672 (guest/guest) - Queue monitoring
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin) - Object storage

### Log Monitoring

```bash
# Monitor all service logs
make monitor-logs

# Monitor individual service logs
make monitor-url-ingestor
make monitor-image-fetcher
make monitor-image-metadata
```

## Testing

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage
```

### Integration Tests

```bash
# Run integration tests
make test-integration
```

The integration test script:
- Checks service health
- Submits test images for processing
- Monitors queue status
- Verifies metrics collection
- Tests the complete processing pipeline

### Test Coverage

After running `make test-coverage`, open `coverage.html` in your browser to view detailed coverage reports.

## API Endpoints

### url-ingestor (Port 8080)

#### Processing Endpoints
- `POST /submit` - Submit image URLs for processing
  - Body: `{"urls": ["http://example.com/image1.jpg", "http://example.com/image2.jpg"]}`

#### Monitoring Endpoints
- `GET /health` - Service health check
- `GET /status` - Service status and dependencies
- `GET /queue/status` - Queue information
- `GET /stats` - System statistics
- `GET /metrics` - Prometheus metrics

### image-fetcher (Port 8081)
- `GET /health` - Service health check
- `GET /metrics` - Prometheus metrics

### image-metadata (Port 8082)
- `GET /health` - Service health check
- `GET /metrics` - Prometheus metrics

## Message Flow

1. Client submits image URLs to url-ingestor
2. url-ingestor publishes messages to RabbitMQ queue "image.urls"
3. image-fetcher consumes messages, downloads images, processes them, and uploads to MinIO
4. image-fetcher publishes results to RabbitMQ queue "image.processed"
5. image-metadata consumes processed messages and stores metadata in PostgreSQL

## Metrics & Monitoring

### Key Metrics

**url-ingestor:**
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `images_submitted_total` - Total images submitted

**image-fetcher:**
- `images_processed_total` - Total images processed (success/error)
- `image_processing_duration_seconds` - Processing time by step
- `active_workers` - Number of active workers
- `queue_size` - Current queue size

**image-metadata:**
- `records_stored_total` - Total records stored (success/error)
- `storage_duration_seconds` - Database operation duration
- `db_connections_active` - Active database connections

### Tracing

The system uses OpenTelemetry with Jaeger for distributed tracing:
- Trace requests across all services
- Monitor processing time for each step
- Debug issues in the processing pipeline

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
├── monitoring/                 # Monitoring configuration
│   ├── prometheus.yml          # Prometheus config
│   └── grafana/                # Grafana dashboards and datasources
├── test/                       # Test files
│   └── integration_test.sh     # Integration test script
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

## Troubleshooting

### Common Issues

1. **Services not starting**: Check dependencies are healthy
   ```bash
   make health-check
   ```

2. **Queue not processing**: Check RabbitMQ management UI
   ```bash
   open http://localhost:15672
   ```

3. **Images not uploading**: Check MinIO console
   ```bash
   open http://localhost:9001
   ```

4. **Database issues**: Check PostgreSQL logs
   ```bash
   docker-compose -f docker-compose.dev.yml logs postgres
   ```

### Debug Commands

```bash
# Check all service status
make status-check

# Monitor logs in real-time
make monitor-logs

# Run integration tests
make test-integration

# Check metrics
make metrics-check
``` 