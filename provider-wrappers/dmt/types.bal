import ballerina/graphql.subgraph;
import ballerina/log;

public type VehicleClass record {|
    readonly string id;
    string className;
|};

@subgraph:Entity {
    key: "nic",
    resolveReference: resolvePersonData
}
public type PersonData record {|
    readonly string nic;
    VehicleInfo[] vehicles;
    DriverLicense? license;
|};

isolated function resolvePersonData(subgraph:Representation representation) returns PersonData|error? {
    string ownerNic = check representation["nic"].ensureType();

    PersonData filteredVehicles = {nic: check representation["nic"].ensureType(), vehicles: [], license: null};


    VehicleInfoResponse|error ownedVehiclesResponse = sharedDMTClient.getVehicles(ownerNic, null, 0, 100);

    if ownedVehiclesResponse is error {
        return filteredVehicles;
    }

    filteredVehicles.vehicles = ownedVehiclesResponse.data;

    DriverLicense|error ownedLicenseResponse = sharedDMTClient.getDriverLicensesByOwnerNic(ownerNic);


    if ownedLicenseResponse is error {
        log:printError("Failed to fetch driver license for ownerNic: ", ownerNic = ownedLicenseResponse.message());
        return filteredVehicles;
    }
    log:printInfo("Resolving driver license data for ownerNic: ", ownedLicenseResponse=ownedLicenseResponse);

    filteredVehicles.license = ownedLicenseResponse;
    
    return filteredVehicles;
}

@subgraph:Entity {
    key: "id ownerNic"
}
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
    string vehicleclassId?;
|};

public type OwnerInfo record {|
    readonly string ownerNic;
    string name;
    string address;
    string birthDate;
    string signatureUrl;
    string bloodGroup;
|};

public type IssuerInfo record {|
    readonly string id;
    string name;
    string issuingAuthority;
    string signatureUrl;
|};

public type Permission record {|
    readonly string id;
    string vehicleType;
    string issueDate;
    string expiryDate;
|};

@subgraph:Entity {
    'key: "id ownerNic"
}
public type DriverLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string frontImageUrl?;
    string backImageUrl?;
    Permission[] permissions;
    OwnerInfo ownerInfo;
    IssuerInfo issuerInfo;
|};


public type VehicleClassResponse record {|
    VehicleClass[] data;
|};

public type PaginationInfo record {|
    int page;
    int pageSize;
    int total;
|};

public type VehicleInfoResponse record {|
    VehicleInfo[] data;
    PaginationInfo pagination;
|};