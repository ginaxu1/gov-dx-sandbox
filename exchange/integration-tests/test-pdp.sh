#!/bin/bash
# Policy Decision Point (PDP) Test Suite

echo "=== Policy Decision Point (PDP) Test Suite ==="
echo ""

# Test function
test_pdp() {
    local test_name="$1"
    local expected="$2"
    local data="$3"
    
    echo "Test: $test_name"
    echo "Expected: $expected"
    echo ""
    
    curl -X POST http://localhost:8082/decide \
      -H "Content-Type: application/json" \
      -d "$data" | jq '.'
    
    echo "---"
}

# Test 1: Valid request with no consent required
test_pdp "Valid request with no consent required" "allow=true, consent_required=false" '{
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
  }'

# Test 2: Valid request with consent required
test_pdp "Valid request with consent required" "allow=true, consent_required=true" '{
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
  }'

# Test 3: Invalid consumer
test_pdp "Invalid consumer" "allow=false, deny_reason=Consumer not found" '{
    "consumer": {
      "id": "unknown-app",
      "name": "Unknown Application"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName"]
    }
  }'

# Test 4: Unauthorized field access
test_pdp "Unauthorized field access" "allow=true, consent_required=true" '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.fullName", "person.birthDate"]
    }
  }'

# Test 5: Invalid action
test_pdp "Invalid action" "allow=false, deny_reason=Invalid action" '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "write",
      "data_fields": ["person.fullName"]
    }
  }'

# Test 6: Single field test
test_pdp "Single field test" "allow=true, consent_required=false" '{
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
  }'

# Test 7: Two fields test
test_pdp "Two fields test" "allow=true, consent_required=false" '{
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
  }'

# Test 8: Mixed fields test
test_pdp "Mixed fields test" "allow=false, deny_reason=not authorized" '{
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
  }'

# Test 9: All approved fields test
test_pdp "All approved fields test" "allow=true, consent_required=false" '{
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
  }'

# Test 10: Single unauthorized field test
test_pdp "Single unauthorized field test" "allow=true, consent_required=true" '{
    "consumer": {
      "id": "passport-app",
      "name": "Passport Application Service",
      "type": "government_service"
    },
    "request": {
      "resource": "person_data",
      "action": "read",
      "data_fields": ["person.birthDate"]
    }
  }'

echo ""
echo "PDP Test Suite Complete"