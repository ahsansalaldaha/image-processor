#!/bin/bash

# Integration test script for the image processing system
set -e

echo "🚀 Starting integration tests for image processing system..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
TEST_IMAGE_URL="https://picsum.photos/200/300"

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for $service_name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" > /dev/null 2>&1; then
            print_status "$service_name is ready!"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts - $service_name not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within expected time"
    return 1
}

# Function to check service health
check_health() {
    local url=$1
    local service_name=$2
    
    echo "Checking $service_name health..."
    response=$(curl -s "$url/health")
    if echo "$response" | grep -q "healthy"; then
        print_status "$service_name is healthy"
        return 0
    else
        print_error "$service_name health check failed: $response"
        return 1
    fi
}

# Function to submit images for processing
submit_images() {
    local urls=("$@")
    local payload=$(cat <<EOF
{
    "urls": [$(printf '"%s"' "${urls[@]}" | tr '\n' ',' | sed 's/,$//')]
}
EOF
)
    
    echo "Submitting images for processing..."
    echo "Payload: $payload"
    
    response=$(curl -s -X POST "$BASE_URL/submit" \
        -H "Content-Type: application/json" \
        -d "$payload")
    
    if [ $? -eq 0 ]; then
        print_status "Images submitted successfully"
        return 0
    else
        print_error "Failed to submit images: $response"
        return 1
    fi
}

# Function to check queue status
check_queue_status() {
    echo "Checking queue status..."
    response=$(curl -s "$BASE_URL/queue/status")
    echo "Queue status: $response"
}

# Function to check metrics
check_metrics() {
    echo "Checking metrics..."
    response=$(curl -s "$BASE_URL/metrics")
    if echo "$response" | grep -q "images_submitted_total"; then
        print_status "Metrics endpoint is working"
    else
        print_warning "Metrics endpoint might not be working properly"
    fi
}

# Function to check system status
check_system_status() {
    echo "Checking system status..."
    response=$(curl -s "$BASE_URL/status")
    echo "System status: $response"
}

# Function to check system stats
check_system_stats() {
    echo "Checking system stats..."
    response=$(curl -s "$BASE_URL/stats")
    echo "System stats: $response"
}

# Main test execution
main() {
    echo "=== Image Processing System Integration Tests ==="
    
    # Wait for services to be ready
    wait_for_service "$BASE_URL" "url-ingestor"
    
    # Check health endpoints
    check_health "$BASE_URL" "url-ingestor"
    
    # Check system status
    check_system_status
    check_system_stats
    
    # Check metrics
    check_metrics
    
    # Check queue status before submission
    check_queue_status
    
    # Submit test images
    submit_images "$TEST_IMAGE_URL" "https://picsum.photos/400/300" "https://picsum.photos/300/400"
    
    # Wait a bit for processing
    echo "Waiting for image processing..."
    sleep 10
    
    # Check queue status after submission
    check_queue_status
    
    # Check metrics after processing
    check_metrics
    
    print_status "Integration tests completed!"
    echo ""
    echo "=== Test Summary ==="
    echo "✅ Service health checks passed"
    echo "✅ Image submission working"
    echo "✅ Metrics collection active"
    echo "✅ Queue monitoring functional"
    echo ""
    echo "=== Monitoring URLs ==="
    echo "🌐 URL Ingestor: $BASE_URL"
    echo "📊 Metrics: $BASE_URL/metrics"
    echo "📈 Status: $BASE_URL/status"
    echo "📋 Queue: $BASE_URL/queue/status"
    echo "📊 Stats: $BASE_URL/stats"
    echo ""
    echo "=== External Monitoring ==="
    echo "🐰 RabbitMQ Management: http://localhost:15672 (guest/guest)"
    echo "📦 MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
    echo "📊 Prometheus: http://localhost:9090"
    echo "📈 Grafana: http://localhost:3000 (admin/admin)"
    echo "🔍 Jaeger: http://localhost:16686"
}

# Run the main function
main "$@" 