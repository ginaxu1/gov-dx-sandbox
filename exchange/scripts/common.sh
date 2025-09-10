#!/bin/bash
# Common configuration and functions for Exchange Services scripts

# Service configuration
PDP_PORT=8082
CE_PORT=8081
PDP_URL="http://localhost:${PDP_PORT}"
CE_URL="http://localhost:${CE_PORT}"
PDP_HEALTH="${PDP_URL}/health"
CE_HEALTH="${CE_URL}/health"

# Common functions
check_docker() {
    docker info > /dev/null 2>&1 || { echo "❌ Docker not running"; exit 1; }
}

wait_for_services() {
    local wait_time=${1:-5}
    echo "Waiting ${wait_time}s for services to start..."
    sleep $wait_time
}

check_health() {
    echo "Health checks:"
    curl -s "$PDP_HEALTH" > /dev/null && echo "✅ PDP (${PDP_PORT})" || echo "❌ PDP"
    curl -s "$CE_HEALTH" > /dev/null && echo "✅ CE (${CE_PORT})" || echo "❌ CE"
}

show_endpoints() {
    local env=${1:-"Local"}
    echo ""
    echo "Endpoints (${env}):"
    echo "   PDP: ${PDP_URL}"
    echo "   CE:  ${CE_URL}"
    echo ""
    echo "Commands: ./scripts/logs.sh | ./scripts/stop.sh | ./scripts/test.sh"
}

# Test data
PDP_TEST_DATA='{"consumer":{"id":"test-app","name":"Test App","type":"mobile_app"},"request":{"resource":"person_data","action":"read","data_fields":["person.fullName"]},"timestamp":"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}'
CE_TEST_DATA='{"consumer_id":"test-app","data_owner":"test-owner","data_fields":["person.fullName"],"purpose":"testing","expiry_days":30}'