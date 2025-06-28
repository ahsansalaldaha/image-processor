#!/bin/bash

# Setup script for Image Processing System
# This script helps set up the development environment

set -e

echo "🚀 Setting up Image Processing System..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p docker/minio
mkdir -p docker/db

# Copy environment file if it doesn't exist
if [ ! -f .env.dev ]; then
    echo "📝 Creating development environment file..."
    cp env.dev.example .env.dev
    echo "✅ Created .env.dev (you can customize it if needed)"
else
    echo "✅ .env.dev already exists"
fi

# Build the development images
echo "🔨 Building development images..."
docker-compose -f docker-compose.dev.yml build

echo ""
echo "✅ Setup complete!"
echo ""
echo "To start development environment:"
echo "  make dev"
echo ""
echo "To start production environment:"
echo "  make prod"
echo ""
echo "To see all available commands:"
echo "  make help"
echo ""
echo "Happy coding! 🎉" 