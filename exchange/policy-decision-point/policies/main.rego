package opendif.authz


# By default, the decision is to deny
default decision = {
    "allow": false,
    "deny_reason": "Not authorized by default policy",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
}

# Main decision rule - allows access if all ABAC conditions are met
decision = {
    "allow": true,
    "deny_reason": null,
    "consent_required": consent_required,
    "consent_required_fields": consent_fields,
    "data_owner": data_owner,
    "expiry_time": expiry_time,
    "conditions": conditions
} {
    # Check if the request meets all ABAC authorization criteria
    abac_authorization_passed
    
    # Determine consent requirements
    consent_fields := get_consent_required_fields(input.request.data_fields)
    consent_required := count(consent_fields) > 0
    
    # Set conditions for the authorization
    conditions := {
        "consumer_verified": true,
        "resource_authorized": true,
        "action_authorized": true
    }
    
    # Get data owner and expiry information for consent fields (only if consent is required)
    data_owner := get_data_owner(consent_fields)
    expiry_time := get_expiry_time(consent_fields)
}

# Simplified ABAC Authorization Rule - focuses on core consent flow requirements
abac_authorization_passed {
    # Subject (Consumer) attributes check
    consumer_authorized
    
    # Resource attributes check
    resource_authorized
    
    # Action attributes check
    action_authorized
}

# Consumer Authorization - checks if the consumer is authorized to access the resource
consumer_authorized {
    # Access consumer data loaded via embedded data module
    consumer_data := consumer_grants[input.consumer.id]
    
    # Check if consumer has access to the requested resource
    resource := input.request.resource
    approved_fields := consumer_data.approved_fields
    requested_fields := input.request.data_fields
    
    # All requested fields must be in approved fields
    all_fields_approved(requested_fields, approved_fields)
}

# Resource Authorization - checks if the resource is accessible
resource_authorized {
    # Check if the resource exists in provider metadata
    resource := input.request.resource
    field := input.request.data_fields[_]
    provider_metadata.fields[field]
}

# Action Authorization - checks if the action is permitted
action_authorized {
    # Only "read" action is currently supported
    input.request.action == "read"
}

# Helper function to check if all requested fields are approved
all_fields_approved(requested_fields, approved_fields) {
    # Convert both lists to sets for efficient comparison
    requested_set := {field | field := requested_fields[_]}
    approved_set := {field | field := approved_fields[_]}
    
    # Check if the requested set is a subset of the approved set
    requested_set <= approved_set
}

# Function to get fields that require consent based on metadata
get_consent_required_fields(requested_fields) = fields {
    fields := [field | 
        field := requested_fields[_]
        provider_metadata.fields[field].consent_required == true
    ]
}

# Function to get the primary data owner for consent fields
get_primary_data_owner(consent_fields) = owner {
    count(consent_fields) > 0
    # Get the first data owner from consent fields
    owner := provider_metadata.fields[consent_fields[0]].owner
}

# Function to get consent expiry time
get_consent_expiry_time(consent_fields) = expiry {
    count(consent_fields) > 0
    # Get expiry time from the first consent field
    expiry := provider_metadata.fields[consent_fields[0]].expiry_time
}

# Helper function to get data owner (returns empty string if no consent fields)
get_data_owner(consent_fields) = owner {
    owner := get_primary_data_owner(consent_fields)
} else = "" {
    count(consent_fields) == 0
}

# Helper function to get expiry time (returns empty string if no consent fields)
get_expiry_time(consent_fields) = expiry {
    expiry := get_consent_expiry_time(consent_fields)
} else = "" {
    count(consent_fields) == 0
}

# Denial rules for specific scenarios
decision = {
    "allow": false,
    "deny_reason": "Consumer not found in grants",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
} {
    not consumer_grants[input.consumer.id]
}

decision = {
    "allow": false,
    "deny_reason": "Consumer not authorized for requested fields",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
} {
    consumer_grants[input.consumer.id]
    not all_fields_approved(input.request.data_fields, consumer_grants[input.consumer.id].approved_fields)
}

decision = {
    "allow": false,
    "deny_reason": "Invalid action requested",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
} {
    input.request.action != "read"
}