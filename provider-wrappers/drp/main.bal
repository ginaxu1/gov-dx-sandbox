import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

// ---  Enum Definitions based on the DRP schema ---
public enum Gender {
    MALE,
    FEMALE
}

public enum CardStatus {
    ACTIVE,
    EXPIRED,
    LOST,
    CANCELLED
}

public enum CivilStatus {
    MARRIED,
    SINGLE,
    WIDOWED,
    DIVORCED
}

public enum CitizenshipType {
    DESCENT,
    REGISTRATION,
    NATURALIZATION
}

// --- Record type definitions based on the DRP Schema ---
// Record type for Card Information
type CardInfo record {|
    readonly string cardNumber;
    string issueDate;
    string expiryDate;
    CardStatus cardStatus;
|};

// Record type for Lost Card Replacement Information
type LostCardReplacementInfo record {|
    string policeStation;
    string complaintDate;
    string complaintNumber;
|};

// Record type for Citizenship Information
type CitizenshipInfo record {|
    CitizenshipType citizenshipType;
    string certificateNumber;
    string issueDate;
|};

// Record type for Parent Information
type ParentInfo record {|
    string fatherName;
    string motherName;
    string fatherNic;
    string motherNic;
|};

// The main entity for the subgraph. This represents a person.
@subgraph:Entity {
    key: ["nic"],
    resolveReference: resolvePersonReference
}
public type PersonInfo record {|
    readonly string nic;
    string fullName;
    string surname;
    string otherNames;
    Gender gender;
    string dateOfBirth;
    string placeOfBirth;
    string permanentAddress;
    string profession;
    CivilStatus civilStatus;
    string contactNumber;
    string email;
    string photo;
|};

// Combined record that includes all related information.
// This will be the return type for direct queries to this subgraph.
type PersonData record {|
    *PersonInfo;
    CardInfo cardInfo;
    LostCardReplacementInfo? lostCardReplacementInfo;
    CitizenshipInfo citizenshipInfo;
    ParentInfo parentInfo;
|};

// --- DRPAPIClient (Provider Wrapper) ---
// This client encapsulates the logic for interacting with the external DRP API service.
isolated service class DRPAPIClient {
    private final http:Client apiClient;

    function init(string baseUrl) returns http:ClientError? {
        self.apiClient = check new (baseUrl);
    }

    isolated function getPersonByNic(string nic) returns PersonData|error {
        log:printInfo("DRPAPIClient: Fetching person from external API", nic = nic);
        
        // In a real scenario, this would be a network call.
        // For this example, we are using the mock data.
        if nic == "199512345678" {
             PersonData personData = {
                nic: "199512345678",
                fullName: "Nuwan Fernando",
                surname: "Fernando",
                otherNames: "Nuwan",
                gender: MALE, // Using enum member
                dateOfBirth: "1995-12-01",
                placeOfBirth: "Colombo",
                permanentAddress: "105 Bauddhaloka Mawatha, Colombo 00400",
                profession: "Software Engineer",
                civilStatus: MARRIED, // Using enum member
                contactNumber: "+94771234567",
                email: "nuwan@opensource.lk",
                photo: "https://example.com/photo.jpg",
                cardInfo: {
                    cardNumber: "199512345678",
                    issueDate: "2018-01-02",
                    expiryDate: "2028-01-01",
                    cardStatus: ACTIVE // Using enum member
                },
                lostCardReplacementInfo: (), // No lost card info for this user
                citizenshipInfo: {
                    citizenshipType: DESCENT, // Using enum member
                    certificateNumber: "A12345",
                    issueDate: "1995-12-02"
                },
                parentInfo: {
                    fatherName: "Father Fernando",
                    motherName: "Ruby de Silva",
                    fatherNic: "196618234567",
                    motherNic: "196817654321"
                }
            };
            return personData;
        }
        return error("Person not found");
    }
}

// Create a shared instance of DRPAPIClient to be used across the subgraph.
final DRPAPIClient sharedDRPClient = check new ("http://localhost:8080");

// Top-Level GraphQL Resolver for Federated References
isolated function resolvePersonReference(map<anydata> representation) returns PersonInfo? {
    // Safely extract the NIC from the representation
    string nic = <string>representation["nic"];
    log:printInfo("DRP Service: Resolving reference for Person via API Client", nic = nic);

    PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
    if personData is error {
        // Centralized error logging for any failure
        log:printWarn("DRP Service: Failed to resolve person reference", nic = nic, err = personData.toString());
        return ();
    }
    PersonInfo|error personInfo = personData.cloneWithType(PersonInfo);
    if personInfo is error {
        log:printWarn("DRP Service: Failed to clone person data to PersonInfo", nic = nic, err = personInfo.toString());
        return ();
    }
    log:printInfo("DRP Service: Successfully resolved reference for Person", nic = nic);
    return personInfo;
}

// --- GraphQL Subgraph Service ---
@subgraph:Subgraph
isolated service / on new graphql:Listener(9091) {

    // Resource function to query a person's full data directly from this subgraph.
    resource function get person(string nic) returns PersonData? {
        log:printInfo("DRP Service: Looking for person via API Client", nic = nic);
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            log:printWarn("DRP Service: Person not found or error fetching person", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }

    // Health check endpoint
    resource function get health() returns string {
        log:printInfo("DRP Service: Health check requested.");
        return "OK";
    }
}