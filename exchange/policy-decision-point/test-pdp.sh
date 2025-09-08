#!/bin/bash

# Test script for Policy Decision Point
# This script demonstrates the ABAC authorization flow

echo "=== Policy Decision Point (PDP) Test Suite ==="
echo ""

# Test 1: Valid request with no consent required
echo "Test 1: Valid request with no consent required"
echo "Requesting: person.fullName, person.nic, person.photo"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic", "person.photo"]
    }
  }' | jq '.'

echo "---"

# Test 2: Valid request with consent required
echo "Test 2: Valid request with consent required"
echo "Requesting: person.fullName, person.permanentAddress, person.birthDate"
echo "Expected: allow=true, consent_required=true, consent_required_fields=[person.permanentAddress, person.birthDate]"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.permanentAddress", "person.birthDate"]
    }
  }' | jq '.'

echo "---"

# Test 3: Invalid consumer
echo "Test 3: Invalid consumer"
echo "Requesting with unknown consumer: unknown-app"
echo "Expected: allow=false, deny_reason=Consumer not found in grants"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "unknown-app",
      "name": "Unknown Application"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName"]
    },
  }' | jq '.'

echo "---"

# Test 4: Unauthorized field access
echo "Test 4: Unauthorized field access"
echo "Requesting unauthorized field: person.ssn"
echo "Expected: allow=false, deny_reason=Consumer not authorized for requested fields"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.ssn"]
    },
  }' | jq '.'

echo "---"

# Test 5: Invalid action
echo "Test 5: Invalid action"
echo "Requesting action: write (not supported)"
echo "Expected: allow=false, deny_reason=Invalid action requested"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service"
    },
    "request": {
      "resource": "person_data",
      "action": "write",
      "data_fields": ["person.fullName"]
    },
  }' | jq '.'

echo "---"

# Test 6: Single field test (person.fullName only)
echo "Test 6: Single field test (person.fullName only)"
echo "Requesting: person.fullName"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName"]
    }
  }' | jq '.'

echo "---"

# Test 7: Two fields test (person.fullName, person.nic)
echo "Test 7: Two fields test (person.fullName, person.nic)"
echo "Requesting: person.fullName, person.nic"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic"]
    }
  }' | jq '.'

echo "---"

# Test 8: Mixed fields test (approved + unapproved)
echo "Test 8: Mixed fields test (approved + unapproved)"
echo "Requesting: person.fullName, person.permanentAddress"
echo "Expected: allow=false, deny_reason=Consumer not authorized for requested fields"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.permanentAddress"]
    }
  }' | jq '.'

echo "---"

# Test 9: All approved fields test
echo "Test 9: All approved fields test"
echo "Requesting: person.fullName, person.nic, person.photo (all approved)"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic", "person.photo"]
    }
  }' | jq '.'

echo "---"

# Test 6: Single field test (person.fullName only)
echo "Test 6: Single field test (person.fullName only)"
echo "Requesting: person.fullName"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName"]
    }
  }' | jq '.'

echo "---"
echo ""

# Test 7: Two fields test (person.fullName, person.nic)
echo "Test 7: Two fields test (person.fullName, person.nic)"
echo "Requesting: person.fullName, person.nic"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic"]
    }
  }' | jq '.'

echo "---"
echo ""

# Test 8: Mixed fields test (approved + unapproved)
echo "Test 8: Mixed fields test (approved + unapproved)"
echo "Requesting: person.fullName, person.permanentAddress"
echo "Expected: allow=false, deny_reason=Consumer not authorized for requested fields"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.permanentAddress"]
    }
  }' | jq '.'

echo "---"
echo ""

# Test 9: All approved fields test
cd /Users/tmp/gov-dx-sandbox/exchange/policy-decision-point && echo "Testing key scenarios:" && echo "" && echo "Test 6: Single field test" && curl -X POST http://localhost:8080/decide -H "Content-Type: application/json" -d '{"consumer":{"id":"passport-app","name":"Passport Application Service","type":"government_service"},"request":{"resource":"person_data","action":"read","data_fields":["person.fullName"]}' | jq '.' && echo "" && echo "Test 1: All approved fields" && curl -X POST http://localhost:8080/decide -H "Content-Type: application/json" -d '{"consumer":{"id":"passport-app","name":"Passport Application Service","type":"government_service"},"request":{"resource":"person_data","action":"read","data_fields":["person.fullName","person.nic","person.photo"]}' | jq '.'echo "Test 9: All approved fields test"
echo "Requesting: person.fullName, person.nic, person.photo (all approved)"
echo "Expected: allow=true, consent_required=false"
echo ""

curl -X POST http://localhost:8080/decide \
  -H "Content-Type: application/json" \
  -d '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.nic", "person.photo"]
    }
  }' | jq '.'

echo "---"
echo ""
