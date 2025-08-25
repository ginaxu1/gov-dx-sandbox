// AUTO-GENERATED FILE. DO NOT MODIFY.

// This file is an auto-generated file by Ballerina persistence layer for model.
// It should not be modified by hand.

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

public type DrivingLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string frontImageUrl;
    string backImageUrl;

    string ownerinfoOwnerNic;
    string issuerinfoId;
|};

public type DrivingLicenseOptionalized record {|
    string id?;
    string licenseNumber?;
    string issueDate?;
    string expiryDate?;
    string frontImageUrl?;
    string backImageUrl?;
    string ownerinfoOwnerNic?;
    string issuerinfoId?;
|};

public type DrivingLicenseWithRelations record {|
    *DrivingLicenseOptionalized;
    VehiclePermissionOptionalized[] permissions?;
    OwnerInfoOptionalized ownerInfo?;
    IssuerInfoOptionalized issuerInfo?;
|};

public type DrivingLicenseTargetType typedesc<DrivingLicenseWithRelations>;

public type DrivingLicenseInsert DrivingLicense;

public type DrivingLicenseUpdate record {|
    string licenseNumber?;
    string issueDate?;
    string expiryDate?;
    string frontImageUrl?;
    string backImageUrl?;
    string ownerinfoOwnerNic?;
    string issuerinfoId?;
|};

public type VehicleClass record {|
    readonly string id;
    string className;

|};

public type VehicleClassOptionalized record {|
    string id?;
    string className?;
|};

public type VehicleClassWithRelations record {|
    *VehicleClassOptionalized;
    VehicleInfoOptionalized[] vehicles?;
|};

public type VehicleClassTargetType typedesc<VehicleClassWithRelations>;

public type VehicleClassInsert VehicleClass;

public type VehicleClassUpdate record {|
    string className?;
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
    string vehicleclassId;
|};

public type VehicleInfoOptionalized record {|
    string id?;
    string make?;
    string model?;
    int yearOfManufacture?;
    string ownerNic?;
    string engineNumber?;
    string conditionAndNotes?;
    string registrationNumber?;
    string vehicleclassId?;
|};

public type VehicleInfoWithRelations record {|
    *VehicleInfoOptionalized;
    VehicleClassOptionalized vehicleClass?;
|};

public type VehicleInfoTargetType typedesc<VehicleInfoWithRelations>;

public type VehicleInfoInsert VehicleInfo;

public type VehicleInfoUpdate record {|
    string make?;
    string model?;
    int yearOfManufacture?;
    string ownerNic?;
    string engineNumber?;
    string conditionAndNotes?;
    string registrationNumber?;
    string vehicleclassId?;
|};

public type OwnerInfo record {|
    readonly string ownerNic;
    string name;
    string address;
    string birthDate;
    string signatureUrl;
    BloodGroup bloodGroup;

|};

public type OwnerInfoOptionalized record {|
    string ownerNic?;
    string name?;
    string address?;
    string birthDate?;
    string signatureUrl?;
    BloodGroup bloodGroup?;
|};

public type OwnerInfoWithRelations record {|
    *OwnerInfoOptionalized;
    DrivingLicenseOptionalized drivingLicense?;
|};

public type OwnerInfoTargetType typedesc<OwnerInfoWithRelations>;

public type OwnerInfoInsert OwnerInfo;

public type OwnerInfoUpdate record {|
    string name?;
    string address?;
    string birthDate?;
    string signatureUrl?;
    BloodGroup bloodGroup?;
|};

public type IssuerInfo record {|
    readonly string id;
    string name;
    string issuingAuthority;
    string signatureUrl;

|};

public type IssuerInfoOptionalized record {|
    string id?;
    string name?;
    string issuingAuthority?;
    string signatureUrl?;
|};

public type IssuerInfoWithRelations record {|
    *IssuerInfoOptionalized;
    DrivingLicenseOptionalized[] drivingLicenses?;
|};

public type IssuerInfoTargetType typedesc<IssuerInfoWithRelations>;

public type IssuerInfoInsert IssuerInfo;

public type IssuerInfoUpdate record {|
    string name?;
    string issuingAuthority?;
    string signatureUrl?;
|};

public type VehiclePermission record {|
    readonly string id;
    VehicleType vehicleType;
    string issueDate;
    string expiryDate;
    string drivinglicenseId;
|};

public type VehiclePermissionOptionalized record {|
    string id?;
    VehicleType vehicleType?;
    string issueDate?;
    string expiryDate?;
    string drivinglicenseId?;
|};

public type VehiclePermissionWithRelations record {|
    *VehiclePermissionOptionalized;
    DrivingLicenseOptionalized drivingLicense?;
|};

public type VehiclePermissionTargetType typedesc<VehiclePermissionWithRelations>;

public type VehiclePermissionInsert VehiclePermission;

public type VehiclePermissionUpdate record {|
    VehicleType vehicleType?;
    string issueDate?;
    string expiryDate?;
    string drivinglicenseId?;
|};

