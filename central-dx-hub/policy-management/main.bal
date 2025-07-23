import ballerina/http;

// Defines the structure for our access control policies.
type Policy record {|
    string consumerOrg;
    string providerService;
    string[] allowedFields;
    boolean requiresConsent;
|};

// A module-level map to store policies in memory.
// The key is a string like "Police:DMV-LicenseAPI".
isolated map<Policy> policies = {};

// The main service that runs on port 9090.
service / on new http:Listener(9090) {

    // Handles creating or updating a policy.
    // POST http://localhost:9090/policy
    isolated resource function post policy(@http:Payload Policy newPolicy) returns http:Created|http:InternalServerError {
        // Create a unique key for the policy lookup.
        string policyKey = string `${newPolicy.consumerOrg}:${newPolicy.providerService}`;
        lock {
            // Add the new policy to our in-memory map.
            policies[policyKey] = newPolicy.clone();
        }

        // Return a 201 Created response with a success message.
        return <http:Created>{body: string `Policy created for ${policyKey}`};
    }

    // Handles retrieving a specific policy.
    // GET http://localhost:9090/policy/Police/DMV-LicenseAPI
    isolated resource function get policy/[string consumer]/[string provider]() returns Policy|http:NotFound {
        // Create the key from the path parameters.
        string policyKey = string `${consumer}:${provider}`;

        // Check if the policy exists and return it.
        lock {
            if policies.hasKey(policyKey) {
                return policies.get(policyKey).clone();
            }
        }

        // If not found, return a 404 Not Found response.
        return <http:NotFound>{body: string `Policy not found for ${policyKey}`};
    }
}