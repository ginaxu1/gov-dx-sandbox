#!/bin/bash

# Test script to manually check Redis stream consumer

echo "🔍 Testing Redis Stream Consumer"
echo "==============================="

# Check if stream exists
echo "1. Checking if stream exists..."
redis-cli xlen audit-events

# Check consumer groups
echo "2. Checking consumer groups..."
redis-cli xinfo groups audit-events

# Check pending messages
echo "3. Checking pending messages..."
redis-cli xpending audit-events audit-processors

# Try to manually read from the stream
echo "4. Trying to manually read from stream..."
redis-cli xreadgroup GROUP audit-processors test-consumer COUNT 1 STREAMS audit-events ">"

# Check stream length after manual read
echo "5. Checking stream length after manual read..."
redis-cli xlen audit-events

# Check pending messages after manual read
echo "6. Checking pending messages after manual read..."
redis-cli xpending audit-events audit-processors

echo "Test completed."
