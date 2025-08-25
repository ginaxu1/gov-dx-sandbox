public enum BloodGroup {
    A_POSITIVE,
    A_NEGATIVE,
    B_POSITIVE,
    B_NEGATIVE,
    AB_POSITIVE,
    AB_NEGATIVE,
    O_POSITIVE,
    O_NEGATIVE
}

public enum VehicleType {
    A1,
    A,
    B,
    C1,
    C,
    CE,
    D1,
    D,
    DE,
    G1,
    G,
    J
}

public type VehicleClass record {|
    readonly string id;
    string className;
|};

public type VehiclePermission record {|
    readonly string id;
    VehicleType vehicleType;
    string issueDate;
    string expiryDate;
|};

public type VehicleInfo record {|
    readonly string id;
    string make;
    string model;
    int yearOfManufacture;
    string ownerNic;
    string engineNumber;
    string conditionAndNotes;
    string registrationNumber;
    VehicleClass vehicleClass;
|};

public type OwnerInfo record {|
    readonly string ownerNic;
    string name;
    string address;
    string birthDate;
    string signatureUrl;
    BloodGroup bloodGroup;
|};

public type IssuerInfo record {|
    readonly string id;
    string name;
    string issuingAuthority;
    string signatureUrl;
|};

public type DrivingLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string frontImageUrl;
    string backImageUrl;
    VehiclePermission[] permissions;
    OwnerInfo ownerInfo;
    IssuerInfo issuerInfo;
|};

