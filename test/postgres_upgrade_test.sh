#!/bin/bash

# PostgreSQL 17 Upgrade Test Script
set -e

echo "🔍 Testing PostgreSQL 17 upgrade..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

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

# Function to wait for PostgreSQL to be ready
wait_for_postgres() {
    local max_attempts=30
    local attempt=1

    echo "Waiting for PostgreSQL 17 to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if docker exec $(docker ps -q -f "name=postgres") pg_isready -U postgres > /dev/null 2>&1; then
            print_status "PostgreSQL 17 is ready!"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts - PostgreSQL not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    print_error "PostgreSQL failed to start within expected time"
    return 1
}

# Function to check PostgreSQL version
check_postgres_version() {
    echo "Checking PostgreSQL version..."
    version=$(docker exec $(docker ps -q -f "name=postgres") psql -U postgres -t -c "SELECT version();" | head -1 | tr -d ' ')
    
    if echo "$version" | grep -q "PostgreSQL 17"; then
        print_status "PostgreSQL 17 is running correctly"
        echo "Version: $version"
        return 0
    else
        print_error "PostgreSQL version check failed. Expected 17, got: $version"
        return 1
    fi
}

# Function to test database connectivity
test_db_connectivity() {
    echo "Testing database connectivity..."
    
    # Test basic connection
    if docker exec $(docker ps -q -f "name=postgres") psql -U postgres -d images -c "SELECT 1;" > /dev/null 2>&1; then
        print_status "Database connectivity test passed"
        return 0
    else
        print_error "Database connectivity test failed"
        return 1
    fi
}

# Function to test GORM auto-migration
test_gorm_migration() {
    echo "Testing GORM auto-migration..."
    
    # Check if the image_records table exists
    table_exists=$(docker exec $(docker ps -q -f "name=postgres") psql -U postgres -d images -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'image_records');" | tr -d ' ')
    
    if [ "$table_exists" = "t" ]; then
        print_status "GORM auto-migration successful - image_records table exists"
        return 0
    else
        print_error "GORM auto-migration failed - image_records table not found"
        return 1
    fi
}

# Function to test application startup
test_app_startup() {
    echo "Testing application startup with PostgreSQL 17..."
    
    # Start the services
    docker-compose -f docker-compose.dev.yml up -d postgres
    
    # Wait for PostgreSQL
    wait_for_postgres
    
    # Check version
    check_postgres_version
    
    # Test connectivity
    test_db_connectivity
    
    # Test migration
    test_gorm_migration
    
    # Start image-metadata service to test GORM migration
    docker-compose -f docker-compose.dev.yml up -d image-metadata
    
    # Wait a bit for the service to start and run migration
    sleep 10
    
    # Check if service is running
    if docker ps | grep -q "image-metadata"; then
        print_status "Application startup with PostgreSQL 17 successful"
        return 0
    else
        print_error "Application startup with PostgreSQL 17 failed"
        return 1
    fi
}

# Function to clean up
cleanup() {
    echo "Cleaning up test environment..."
    docker-compose -f docker-compose.dev.yml down
}

# Main test execution
main() {
    echo "=== PostgreSQL 17 Upgrade Test ==="
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    # Test application startup
    test_app_startup
    
    print_status "PostgreSQL 17 upgrade test completed successfully!"
    echo ""
    echo "=== Test Summary ==="
    echo "✅ PostgreSQL 17 is running"
    echo "✅ Database connectivity working"
    echo "✅ GORM auto-migration successful"
    echo "✅ Application startup successful"
    echo ""
    echo "=== PostgreSQL 17 Benefits ==="
    echo "🚀 Improved performance and scalability"
    echo "🔒 Enhanced security features"
    echo "📊 Better monitoring and observability"
    echo "🛠️  Latest features and bug fixes"
}

# Run the main function
main "$@" 