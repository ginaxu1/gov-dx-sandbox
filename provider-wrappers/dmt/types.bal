public type VehicleClass record {|
    readonly string id;
    string className;
|};

public type VehicleInfo record {|
    readonly string id;
    string make;
    string model;
    int yearOfManufacture;
    string ownerId;
    string engineNumber;
    string conditionAndNotes;
    string registrationNumber;
    VehicleClass vehicleClass;
|};

public type DriverLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string? photoUrl;
    string ownerId;
|};