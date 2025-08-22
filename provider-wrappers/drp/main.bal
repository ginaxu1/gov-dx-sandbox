import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;
import ballerina/os;

// --- DRPAPIClient (Provider Wrapper) ---
configurable int port = ?;

// Read environment variables
configurable string? serviceURL = ();
configurable string? choreoApiKey = ();

// Use the configurable variable if it exists, otherwise fall back to the environment variable
final string SERVICE_URL = serviceURL ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_SERVICEURL");
final string CHOREO_API_KEY = choreoApiKey ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_APIKEY");

isolated service class DRPAPIClient {
    private final http:Client apiClient;
    function init() returns http:ClientError? {
        log:printInfo("DRPAPIClient: Initializing", apiKey = CHOREO_API_KEY);
        self.apiClient = check new (SERVICE_URL);
    }
    isolated function getPersonByNic(string nic) returns PersonData|error {
        log:printInfo("DRPAPIClient: Fetching person from external API", nic = nic);
        string path = string `/person/${nic}`;
        return self.apiClient->get(path, {"Choreo-API-Key": CHOREO_API_KEY});
    }
}

// This function initializes the DRPAPIClient and is used in the main GraphQL service.
public function initializeDRPClient() returns DRPAPIClient|error {
    return new ();
}

// Shared instance of the DRPAPIClient to be used across the service.
// This is initialized once and used for all requests to avoid creating multiple clients.
final DRPAPIClient sharedDRPClient = check initializeDRPClient();

// --- GraphQL Subgraph Service ---
@subgraph:Subgraph
isolated service / on new graphql:Listener(port) {
    // Fetches the full person data for a given NIC.
    resource function get person(@graphql:ID string nic) returns PersonData? {
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            log:printWarn("DRP Service: Person not found or error fetching person", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }
}