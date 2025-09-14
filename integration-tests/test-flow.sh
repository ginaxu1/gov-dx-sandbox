#!/bin/bash

# Test script to verify the complete flow from Sri Lanka Passport to Policy Decision Point

echo "=== Starting Services for Flow Test ==="

# Kill any existing processes
echo "Killing existing processes..."
pkill -f "consent-engine" || true
pkill -f "orchestration-engine-go" || true
pkill -f "policy-decision-point" || true
sleep 2

# Start Policy Decision Point
echo "Starting Policy Decision Point..."
cd /Users/tmp/gov-dx-sandbox/exchange/policy-decision-point
go run main.go policy-evaluator.go &
PDP_PID=$!
echo "Policy Decision Point started with PID: $PDP_PID"

# Wait for PDP to start
sleep 3

# Test PDP health
echo "Testing Policy Decision Point health..."
curl -s http://localhost:8082/health || echo "PDP not responding"

# Start Consent Engine
echo "Starting Consent Engine..."
cd /Users/tmp/gov-dx-sandbox/exchange/consent-engine
go run main.go engine.go &
CE_PID=$!
echo "Consent Engine started with PID: $CE_PID"

# Wait for CE to start
sleep 3

# Test CE health
echo "Testing Consent Engine health..."
curl -s http://localhost:8081/health || echo "CE not responding"

# Start Orchestration Engine
echo "Starting Orchestration Engine..."
cd /Users/tmp/gov-dx-sandbox/exchange/orchestration-engine-go
go run main.go &
OE_PID=$!
echo "Orchestration Engine started with PID: $OE_PID"

# Wait for OE to start
sleep 3

# Test OE health
echo "Testing Orchestration Engine health..."
curl -s http://localhost:4000/health || echo "OE not responding"

echo ""
echo "=== Testing Complete Flow ==="

# Test the complete flow
echo "Testing POST /getData to Orchestration Engine..."

# Create a test GraphQL request
cat > /tmp/test_request.json << EOF
{
  "query": "query GetData(\$nic: ID!) { person(nic: \$nic) { surname otherNames birthPlace fullName nic permanentAddress photo birthDate father { name } mother { name } birthCertificateNo district sex profession } }",
  "variables": {
    "nic": "199512345678"
  }
}
EOF

# Test the flow
echo "Sending request to orchestration engine..."
curl -X POST http://localhost:4000/getData \
  -H "Content-Type: application/json" \
  -d @/tmp/test_request.json \
  | jq .

echo ""
echo "=== Services Status ==="
echo "Policy Decision Point PID: $PDP_PID"
echo "Consent Engine PID: $CE_PID" 
echo "Orchestration Engine PID: $OE_PID"

echo ""
echo "=== To stop all services, run: ==="
echo "kill $PDP_PID $CE_PID $OE_PID"

# Clean up test file
rm -f /tmp/test_request.json
