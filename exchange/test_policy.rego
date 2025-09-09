package test

import rego.v1

# Test the all_fields_approved function
test_all_fields_approved := all_fields_approved(["person.birthDate"], ["person.fullName", "person.nic", "person.photo"])

# Test data
consumer_grants := {
    "passport-app": {
        "approved_fields": ["person.fullName", "person.nic", "person.photo"]
    }
}

# Helper function to check if all requested fields are approved
all_fields_approved(requested_fields, approved_fields) {
    # Convert both lists to sets for efficient comparison
    requested_set := {field | field := requested_fields[_]}
    approved_set := {field | field := approved_fields[_]}
    
    # Check if the requested set is a subset of the approved set
    requested_set <= approved_set
}
