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
        nic: "nayana@opensource.lk",
        fullName: "Nuwan Fernando",
        otherNames: "Nuwan",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "105 Bauddhaloka Mawatha, Colombo 00400",
        profession: "Software Engineer",
        photo: "https://example.com/photo.jpg"
    },
    {
        nic: "mohamed@opensource.lk",
        fullName: "Mohamed Ali",
        otherNames: "Mohamed",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "10 Sinha Mawatha, Colombo 00400",
        profession: "Pilot",
        photo: "https://example.com/photo.jpg"
    },
    {
        nic: "regina@opensource.lk",
        fullName: "Regina George",
        otherNames: "Regina",
        sex: FEMALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "1034 Sinha Mawatha, Colombo 00400",
        profession: "Army Commander",
        photo: "https://example.com/photo.jpg"
    },
    {
        nic: "thanikan@opensource.lk",
        fullName: "Thanikan Jayasuriya",
        otherNames: "Thanikan",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "1034 Sinha Mawatha, Colombo 00400",
        profession: "Civil Engineer",
        photo: "https://example.com/photo.jpg"
    },
    {
        nic: "sanjiva@opensource.lk",
        fullName: "Sanjiva Edirisinghe",
        otherNames: "Sanjiva",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "14 Anuruddha Mawatha, Colombo 00400",
        profession: "CEO of OSW2",
        photo: "https://example.com/photo.jpg"
    },
    {
        nic: "thushara@opensource.lk",
        fullName: "Thushara Perera",
        otherNames: "Thushara",
        sex: MALE,
        dateOfBirth: "1995-12-01",
        permanentAddress: "14 Araliya Mawatha, Wattala",
        profession: "Politician",
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
