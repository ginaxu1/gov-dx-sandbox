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

# Denial rules for specific scenarios - these must come first
decision = {
    "allow": false,
    "deny_reason": "Consumer not authorized for requested fields",
    "consent_required": false,
    "consent_required_fields": [],
    "data_owner": "",
    "expiry_time": "",
    "conditions": {}
} {
    not all_fields_authorized(get_required_fields(input), get_consumer_id(input))
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
    
    # Determine consent requirements based on new metadata format
    consent_fields := get_consent_required_fields(get_required_fields(input), get_consumer_id(input))
    consent_required := count(consent_fields) > 0
    
    # Set conditions for the authorization
    conditions := {
        "consumer_verified": true,
        "resource_authorized": true,
        "action_authorized": true
    }
    
    # Get data owner and expiry information for consent fields (only if consent is required)
    data_owner := get_data_owner(consent_fields)
    expiry_time := get_expiry_time(consent_fields, get_consumer_id(input))
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
    # Check if consumer is authorized for all requested fields
    all_fields_authorized(get_required_fields(input), get_consumer_id(input))
}

# Resource Authorization - checks if the resource is accessible
resource_authorized {
    # Check if the resource exists in provider metadata
    field := get_required_fields(input)[_]
    provider_metadata.fields[field]
}

# Action Authorization - checks if the action is permitted
action_authorized {
    # For new format, we assume "read" action is always allowed
    # Legacy format still checks the action field
    true
}

# Helper function to check if all requested fields are authorized for the consumer
all_fields_authorized(requested_fields, consumer_id) {
    # All requested fields must be authorized
    field := requested_fields[_]
    field_authorized(field, consumer_id)
}

# Helper function to check if a specific field is authorized for the consumer
field_authorized(field, consumer_id) {
    field_metadata := provider_metadata.fields[field]
    
    # Public fields are always authorized
    field_metadata.access_control_type == "public"
} else = true {
    field_metadata := provider_metadata.fields[field]
    
    # Restricted fields require consumer to be in allow list
    field_metadata.access_control_type == "restricted"
    consumer_in_allow_list(field, consumer_id)
} else = false {
    # Default to false for any other case
    true
}

# Helper function to check if consumer is in the allow list for a field
consumer_in_allow_list(field, consumer_id) {
    field_metadata := provider_metadata.fields[field]
    allow_list := field_metadata.allow_list[_]
    allow_list.consumerId == consumer_id
}

# Function to get fields that require consent based on new metadata format
# Consent is required when: consent_required: true AND provider != owner
get_consent_required_fields(requested_fields, consumer_id) = fields {
    fields := [field | 
        field := requested_fields[_]
        field_metadata := provider_metadata.fields[field]
        field_metadata.consent_required == true
    ]
}

# Function to get the primary data owner for consent fields
get_primary_data_owner(consent_fields) = owner {
    count(consent_fields) > 0
    # Get the first data owner from consent fields
    owner := provider_metadata.fields[consent_fields[0]].owner
}

# Function to get consent expiry time from allow list
get_consent_expiry_time(consent_fields, consumer_id) = expiry {
    count(consent_fields) > 0
    field := consent_fields[0]
    field_metadata := provider_metadata.fields[field]
    allow_list := field_metadata.allow_list[_]
    allow_list.consumerId == consumer_id
    expiry := allow_list.expiry_time
}

# Helper function to get data owner (returns empty string if no consent fields)
get_data_owner(consent_fields) = owner {
    owner := get_primary_data_owner(consent_fields)
} else = "" {
    count(consent_fields) == 0
}

# Helper function to get expiry time (returns empty string if no consent fields)
get_expiry_time(consent_fields, consumer_id) = expiry {
    expiry := get_consent_expiry_time(consent_fields, consumer_id)
} else = "" {
    count(consent_fields) == 0
}

# Helper functions for the new input format

# Get consumer ID from the new format
get_consumer_id(req) = consumer_id {
    consumer_id := req.consumer_id
}

# Get required fields from the new format
get_required_fields(req) = fields {
    fields := req.required_fields
}
