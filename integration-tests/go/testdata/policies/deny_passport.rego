package opendif.authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow := false

# Deny access to passport-related data
allow if {
    not contains(input.requestedData, "passport")
}

# Allow access to non-restricted fields
allow if {
    input.access_control_type == "public"
}

# Allow access to restricted fields if consumer is in allow_list
allow if {
    input.access_control_type == "restricted"
    consumer_id := input.consumer_id
    consumer_id in input.allow_list
}

