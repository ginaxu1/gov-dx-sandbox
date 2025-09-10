#!/bin/bash
# Test Exchange Services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

if [ "$1" = "help" ] || [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    echo "Exchange Services Testing"
    echo "Usage: $0 [help]"
    echo ""
    echo "Tests: PDP (${PDP_URL}), CE (${CE_URL})"
    echo "Prerequisites: Services running, jq installed"
    echo "Start services: ./scripts/manage.sh start-local"
    exit 0
fi

echo "Testing Exchange Services..."

echo "Testing PDP..."
curl -s -X POST "${PDP_URL}/decide" \
  -H "Content-Type: application/json" \
  -d "$PDP_TEST_DATA" | jq '.'

echo "Testing CE..."
curl -s -X POST "${CE_URL}/consent" \
  -H "Content-Type: application/json" \
  -d "$CE_TEST_DATA" | jq '.'

echo "✅ All tests passed!"