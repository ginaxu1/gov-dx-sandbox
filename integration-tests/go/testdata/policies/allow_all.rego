package opendif.authz

import future.keywords.if
import future.keywords.in

# Default deny
default allow := false

# Allow all authenticated requests
allow if {
    input.subject.authenticated == true
}

# Allow requests from authorized consumers
allow if {
    input.consumer_id in data.consumers.authorized
}

# Allow access to public fields
allow if {
    input.field_name in ["person.fullName", "person.birthDate"]
}

