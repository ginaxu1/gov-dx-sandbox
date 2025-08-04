import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

@subgraph:Entity {
    key: ["id"],
    resolveReference: resolveUserReference 
}
public type User record {| 
    readonly string id;
    string name;
    string dateOfBirth;
|};

// --- UserAPIClient (Our "Provider Wrapper") ---
// This client encapsulates the logic for interacting with the external User API service.
isolated service class UserAPIClient {
    private final http:Client apiClient;
    private final string baseUrl;
    function init(string baseUrl) returns http:ClientError? {
        self.baseUrl = baseUrl;
        self.apiClient = check new (baseUrl);
    }
    isolated function getUserById(string userId) returns User|error {
        log:printInfo("UserAPIClient: Fetching user from external API", id = userId);
        string path = string `/users/${userId}`;
        json|error response = self.apiClient->get(path);
        if response is error {
            log:printError("UserAPIClient: Failed to fetch user from external API", err = response.toString());
            return error("Failed to fetch user from external API: " + response.message());
        }
        if (response is json) {
            // Validate the structure received from the external API to prevent runtime errors
            // Ensure that the 'response' JSON has 'id', 'name', 'dateOfBirth' fields.
            if response is map<json> &&
                response.hasKey("id") && response.hasKey("name") && response.hasKey("dateOfBirth") {
                // Attempt to convert the JSON to the User record type.
                // If 'response.cloneWithType()' fails (e.g., due to type mismatch within fields),
                // it will return an error, which the 'check' keyword will propagate.
                record {|string id; string name; string dateOfBirth;|} userRecord = check response.cloneWithType();
                return {id: userRecord.id, name: userRecord.name, dateOfBirth: userRecord.dateOfBirth};
            } else {
                // If the JSON is valid but missing required keys
                return error("UserAPIClient: External API response missing expected fields for User (id, name, dateOfBirth)");
            }
        }
    }
}

// Create a shared instance of UserAPIClient to be used across the subgraph.
final UserAPIClient sharedUserClient = check new ("http://localhost:8080");

// Top-Level GraphQL Resolver for Federated References
isolated function resolveUserReference(map<anydata> representation) returns User? {
    string id = <string>representation["id"];
    log:printInfo("ROP Service: Resolving reference for User via API Client (Top-Level Function)", userId = id);
    // Directly use the globally available 'sharedUserClient'
    User|error user = sharedUserClient.getUserById(id);
    if user is error {
        log:printWarn("ROP Service: User not found or error resolving reference (Top-Level Function)", userId = id, err = user.toString());
        return ();
    }
    return user;
}

// --- GraphQL Schema Types ---
// This defines the public contract of our GraphQL API.
@subgraph:Subgraph
isolated service / on new graphql:Listener(9091) {
    function init() returns error? {
        // No initialization of UserAPIClient needed here, as 'sharedUserClient' is global.
        // Perform any other service-specific initializations if required.
        return;
    }
    resource function get user(string id) returns User? {
        log:printInfo("ROP Service: Looking for user via API Client", id = id);
        User|error user = sharedUserClient.getUserById(id);
        if user is error {
            log:printWarn("ROP Service: User not found or error fetching user", id = id, err = user.toString());
            return ();
        }
        return user;
    }
    resource function get health() returns string {
        log:printInfo("ROP Service: Health check requested.");
        return "OK";
    }
}
