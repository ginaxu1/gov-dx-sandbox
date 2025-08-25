import ballerina/http;
import ballerina/log;

// --- Enum Definitions ---
// These enums match the schema used in the main GraphQL service.
public enum SEX {
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

// --- Record Type Definitions ---
// These records define the structure of the data served by this mock API.
type CardInfo record {|
    readonly string cardNumber;
    string issueDate;
    string expiryDate;
    CardStatus cardStatus;
|};

type LostCardReplacementInfo record {|
    string policeStation;
    string complaintDate;
    string complaintNumber;
|};

type CitizenshipInfo record {|
    CitizenshipType citizenshipType;
    string certificateNumber;
    string issueDate;
|};

type ParentInfo record {|
    string fatherName;
    string motherName;
    string fatherNic;
    string motherNic;
|};

public type PersonInfo record {|
    readonly string nic;
    string fullName;
    string otherNames;
    SEX sex;
    string dateOfBirth;
    string permanentAddress;
    string profession;
    string photo;
|};

// The full data structure that this mock API will return.
type PersonData record {|
    *PersonInfo;
|};

// --- Mock Data Store ---
// This is an in-memory table that simulates a database for the mock API.
isolated final table<PersonData> key(nic) mockPersonDataTable = table [
    {
        nic: "199512345678",
        fullName: "Nuwan Fernando",
        otherNames: "Nuwan",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "105 Bauddhaloka Mawatha, Colombo 00400",
        profession: "Software Engineer",
        photo: "https://example.com/photo.jpg"
    }
];

// --- Mock HTTP Service ---
// This service simulates the actual DRP backend API.
configurable int PORT = ?;
// The main GraphQL service (provider-wrappers/drp/main.bal) will communicate with this.
isolated service / on new http:Listener(PORT) {

    isolated resource function get person/[string nic]() returns PersonData|http:NotFound {
        log:printInfo("Mock DRP API: Request received for person", nic = nic);
        lock {
            // check whether person exists
            if (!mockPersonDataTable.hasKey(nic)) {
                log:printWarn("Mock DRP API: Person not found", nic = nic);
                return http:NOT_FOUND;
            }

            PersonData person = mockPersonDataTable.get(nic);
            log:printInfo("Mock DRP API: Found person, returning data.", nic = nic);
            return person.clone();
        }
    }
}
