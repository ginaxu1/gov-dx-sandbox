#!/bin/bash
# Start Exchange Services in Production Environment

set -e

echo "Starting Exchange Services (Production Environment)..."

# Check Docker
docker info > /dev/null 2>&1 || { echo "❌ Docker not running"; exit 1; }

# Start with production environment
docker compose --env-file .env.production up --build -d

# Wait and check health
sleep 5
echo "Health checks:"
curl -s http://localhost:8082/health > /dev/null && echo "✅ PDP (8082)" || echo "❌ PDP"
curl -s http://localhost:8081/health > /dev/null && echo "✅ CE (8081)" || echo "❌ CE"

echo ""
echo "Endpoints (Production):"
echo "   PDP: http://localhost:8082"
echo "   CE:  http://localhost:8081"
echo ""
echo "Commands: ./scripts/logs.sh | ./scripts/stop.sh | ./scripts/test.sh"