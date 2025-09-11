package opendif.authz

# By default, the decision is to deny
default decision = {
    "allow": false,
    "consent_required": false,
    "consent_required_fields": []
}

# Main decision rule - allows access if all fields are authorized
decision = {
    "allow": true,
    "consent_required": consent_required,
    "consent_required_fields": consent_fields
} {
    # Check if all requested fields are authorized for the app
    all_fields_authorized(get_required_fields(input), get_app_id(input))
    
    # Determine consent requirements
    consent_fields := get_consent_required_fields(get_required_fields(input), get_app_id(input))
    consent_required := count(consent_fields) > 0
}

# Helper function to check if all requested fields are authorized for the app
all_fields_authorized(requested_fields, app_id) {
    # All requested fields must be authorized
    field := requested_fields[_]
    field_authorized(field, app_id)
}

# Helper function to check if a specific field is authorized for the app
field_authorized(field, app_id) {
    field_metadata := provider_metadata.fields[field]
    
    # Public fields are always authorized
    field_metadata.access_control_type == "public"
} else = true {
    field_metadata := provider_metadata.fields[field]
    
    # Restricted fields require app to be in allow list
    field_metadata.access_control_type == "restricted"
    app_in_allow_list(field, app_id)
} else = false {
    # Default to false for any other case
    true
}

# Helper function to check if app is in the allow list for a field
app_in_allow_list(field, app_id) {
    field_metadata := provider_metadata.fields[field]
    allow_list := field_metadata.allow_list[_]
    allow_list.consumerId == app_id
}

# Function to get fields that require consent
# Consent is required when: consent_required: true AND owner != provider
get_consent_required_fields(requested_fields, app_id) = fields {
    fields := [field | 
        field := requested_fields[_]
        field_metadata := provider_metadata.fields[field]
        field_metadata.consent_required == true
        field_metadata.owner != field_metadata.provider
    ]
}


# Helper functions for input format

# Get consumer ID from the input
get_consumer_id(req) = consumer_id {
    consumer_id := req.consumer_id
}

# Get required fields from the input
get_required_fields(req) = fields {
    fields := req.required_fields
}

# Get app ID from the input
get_app_id(req) = app_id {
    app_id := req.app_id
}
