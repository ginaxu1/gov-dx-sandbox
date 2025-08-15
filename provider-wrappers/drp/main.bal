import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;
import ballerina/os;

// --- DRPAPIClient (Provider Wrapper) ---
configurable int port = ?;

// Read environment variables
configurable string? serviceURL = ();
configurable string? consumerKey = ();
configurable string? consumerSecret = ();
configurable string? tokenURL = ();
configurable string? choreoApiKey = ();

// Use the configurable variable if it exists, otherwise fall back to the environment variable
final string SERVICE_URL = serviceURL ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_SERVICEURL");
final string CONSUMER_KEY = consumerKey ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_CONSUMERKEY");
final string CONSUMER_SECRET = consumerSecret ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_CONSUMERSECRET");
final string TOKEN_URL = tokenURL ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_TOKENURL");
final string CHOREO_API_KEY = choreoApiKey ?: os:getEnv("CHOREO_MOCK_DRP_CONNECTION_APIKEY");

isolated service class DRPAPIClient {
    private final http:Client apiClient;

    function init() returns http:ClientError? {
        log:printInfo("DRPAPIClient: Initializing", consumerKey = CONSUMER_KEY, consumerSecret = CONSUMER_SECRET, apiKey = CHOREO_API_KEY);
        self.apiClient = check new (SERVICE_URL,
            auth = {
                tokenUrl: TOKEN_URL,
                clientId: CONSUMER_KEY,
                clientSecret: CONSUMER_SECRET
            }
        );
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
    resource function get person/getPersonByNic(string nic) returns PersonData? {
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            log:printWarn("DRP Service: Person not found or error fetching person", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }

    // Fetches only the card status for a given NIC.
    resource function get cardStatus(string nic) returns CardStatus? {
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            return ();
        }
        return personData.cardInfo.cardStatus;
    }

    // Fetches only the parent information for a given NIC.
    resource function get parentInfo(string nic) returns ParentInfo? {
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            return ();
        }
        return personData.parentInfo;
    }

    // Fetches information about a lost card report, if one exists.
    resource function get lostCardInfo(string nic) returns LostCardReplacementInfo? {
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            return ();
        }
        return personData.lostCardReplacementInfo;
    }

    // Health check endpoint for the DRP service.
    resource function get drp/health() returns string {
        return "OK";
    }
}
