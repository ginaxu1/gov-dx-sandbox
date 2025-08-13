import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;
import ballerina/os;

// --- DRPAPIClient (Provider Wrapper) ---
configurable int port = ?;

// Read environment variables
configurable string serviceURL = ?;
configurable string consumerKey = ?;
configurable string consumerSecret = ?;
configurable string tokenURL = ?;
configurable string choreoApiKey = ?;

// print the consumerKey and consumerSecret


isolated service class DRPAPIClient {
    private final http:Client apiClient;
    function init() returns http:ClientError? {
        log:printInfo("DRPAPIClient: Initializing", consumerKey = consumerKey, consumerSecret = consumerSecret, apiKey = choreoApiKey);
        self.apiClient = check new (serviceURL,
        auth = {
            tokenUrl: tokenURL,
            clientId: consumerKey,
            clientSecret: consumerSecret
        });
    }
    isolated function getPersonByNic(string nic) returns PersonData|error {
        log:printInfo("DRPAPIClient: Fetching person from external API", nic = nic);
        string path = string `/person/${nic}`;
        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
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
    resource function get person/ getPersonByNic(string nic) returns PersonData? {
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
    resource function get drp/ health() returns string {
        return "OK";
    }
}
