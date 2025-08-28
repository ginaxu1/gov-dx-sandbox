package opendif.authz

# By default, the decision is to deny
default decision = {
    "allow": false,
    "deny_reason": "Not authorized by default policy",
    "consent_required_fields": []
}

decision = {
    "allow": true,
    "deny_reason": null,
    "consent_required_fields": []
} if true