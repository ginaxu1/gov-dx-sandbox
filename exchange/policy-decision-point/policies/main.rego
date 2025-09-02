package opendif.authz

# By default, the decision is to deny
default decision = {
    "allow": false,
    "deny_reason": "Not authorized by default policy",
    "consent_required_fields": []
}

# TODO: update logic, currently always returns true until this is implemented
decision = {
    "allow": true,
    "deny_reason": null,
    "consent_required_fields": []
} {
    # The body of the rule must be inside curly braces
    true
}