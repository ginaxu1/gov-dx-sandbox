// This file centralizes all the data structures for the DRP service.

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
public type CardInfo record {|
    readonly string cardNumber;
    string issueDate;
    string expiryDate;
    CardStatus cardStatus;
|};

public type LostCardReplacementInfo record {|
    string policeStation;
    string complaintDate;
    string complaintNumber;
|};

public type CitizenshipInfo record {|
    CitizenshipType citizenshipType;
    string certificateNumber;
    string issueDate;
|};

public type ParentInfo record {|
    string fatherName;
    string motherName;
    string fatherNic;
    string motherNic;
|};

// This is the main entity for the subgraph.
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

// This is the combined record for the full data set.
public type PersonData record {|
    *PersonInfo;
    CardInfo cardInfo;
    LostCardReplacementInfo? lostCardReplacementInfo;
    CitizenshipInfo citizenshipInfo;
    ParentInfo parentInfo;
|};
