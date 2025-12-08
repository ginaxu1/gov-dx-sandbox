#!/bin/bash

# Script to start Grafana and Prometheus locally

set -e

echo "=========================================="
echo "Starting Observability Stack"
echo "=========================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ ERROR: Docker daemon is not running"
    echo ""
    echo "Please start Docker Desktop and try again."
    exit 1
fi

echo "✓ Docker is running"
echo ""

# Navigate to observability directory
cd "$(dirname "$0")"

# Check if opendif-network exists, create if not
# This network is shared with other services (exchange services, etc.)
if ! docker network ls | grep -q opendif-network; then
    echo "Creating opendif-network..."
    docker network create opendif-network
    echo "✓ Network created"
else
    echo "✓ opendif-network exists"
fi
echo ""

# Start services
echo "Starting Prometheus and Grafana..."
docker compose up -d

echo ""
echo "=========================================="
echo "Services Started!"
echo "=========================================="
echo ""
echo "📊 Grafana Dashboard:"
echo "   http://localhost:3002"
echo "   Login: admin / admin"
echo ""
echo "📈 Prometheus:"
echo "   http://localhost:9091"
echo ""
echo "To view logs:"
echo "   docker compose logs -f"
echo ""
echo "To stop services:"
echo "   docker compose down"
echo ""
echo "=========================================="
echo ""
echo "Checking service status..."
sleep 2
docker compose ps



