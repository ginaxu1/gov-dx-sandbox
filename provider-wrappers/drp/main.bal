import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

// ---  Enum Definitions based on the DRP schema ---
public enum Gender { MALE, FEMALE }
public enum CardStatus { ACTIVE, EXPIRED, LOST, CANCELLED }
public enum CivilStatus { MARRIED, SINGLE, WIDOWED, DIVORCED }
public enum CitizenshipType { DESCENT, REGISTRATION, NATURALIZATION }

// --- Record type definitions based on the DRP Schema ---
type CardInfo record {| readonly string cardNumber; string issueDate; string expiryDate; CardStatus cardStatus; |};
type LostCardReplacementInfo record {| string policeStation; string complaintDate; string complaintNumber; |};
type CitizenshipInfo record {| CitizenshipType citizenshipType; string certificateNumber; string issueDate; |};
type ParentInfo record {| string fatherName; string motherName; string fatherNic; string motherNic; |};
@subgraph:Entity { key: ["nic"], resolveReference: resolvePersonReference }
public type PersonInfo record {| readonly string nic; string fullName; string surname; string otherNames; Gender gender; string dateOfBirth; string placeOfBirth; string permanentAddress; string profession; CivilStatus civilStatus; string contactNumber; string email; string photo; |};
type PersonData record {| *PersonInfo; CardInfo cardInfo; LostCardReplacementInfo? lostCardReplacementInfo; CitizenshipInfo citizenshipInfo; ParentInfo parentInfo; |};

// --- DRPAPIClient (Provider Wrapper) ---
// This client now makes a real HTTP call to the backend service.
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

// This function is now public, making it visible to the test file for mocking.
public function initializeDRPClient() returns DRPAPIClient|error {
    return new ("http://localhost:8080");
}

// The final client now calls the initialization function.
final DRPAPIClient sharedDRPClient = check initializeDRPClient();

// Top-Level GraphQL Resolver for Federated References
isolated function resolvePersonReference(map<anydata> representation) returns PersonInfo? {
    string nic = <string>representation["nic"];
    log:printInfo("DRP Service: Resolving reference for Person", nic = nic);
    PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
    if personData is error {
        log:printWarn("DRP Service: Failed to resolve person reference", nic = nic, err = personData.toString());
        return ();
    }
    // Correctly handle the potential error from cloneWithType
    PersonInfo|error personInfo = personData.cloneWithType(PersonInfo);
    if personInfo is error {
        log:printWarn("DRP Service: Failed to clone PersonData to PersonInfo", nic = nic, err = personInfo.toString());
        return ();
    }
    return personInfo;
}

// --- GraphQL Subgraph Service ---
@subgraph:Subgraph
isolated service / on new graphql:Listener(9091) {
    resource function get person(string nic) returns PersonData? {
        log:printInfo("DRP Service: Looking for person via API Client", nic = nic);
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            log:printWarn("DRP Service: Person not found or error fetching person", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }
    resource function get health() returns string {
        log:printInfo("DRP Service: Health check requested.");
        return "OK";
    }
}
