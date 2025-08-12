import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

configurable int PORT = ?;
configurable string DRP_API_BASE_URL = ?;

// --- DRPAPIClient (Provider Wrapper) ---
// This client makes a real HTTP call to the backend service.
isolated service class DRPAPIClient {
    private final http:Client apiClient;
    function init(string baseUrl) returns http:ClientError? {
        self.apiClient = check new (baseUrl);
    }
    isolated function getPersonByNic(string nic) returns PersonData|error {
        log:printInfo("DRPAPIClient: Fetching person from external API", nic = nic);
        string path = string `/person/${nic}`;
        return self.apiClient->get(path);
    }
}

// This function initializes the DRPAPIClient and is used in the main GraphQL service.
public function initializeDRPClient() returns DRPAPIClient|error {
    return new (DRP_API_BASE_URL);
}

// Shared instance of the DRPAPIClient to be used across the service.
// This is initialized once and used for all requests to avoid creating multiple clients.
final DRPAPIClient sharedDRPClient = check initializeDRPClient();

// --- GraphQL Subgraph Service ---
@subgraph:Subgraph
isolated service / on new graphql:Listener(PORT) {
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
