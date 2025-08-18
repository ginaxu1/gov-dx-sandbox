import ballerina/graphql.subgraph;

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

    VehicleInfo[] ownedVehicles;
    lock {
        VehicleInfo[] tempVehicles = [];
        foreach var vehicle in vehicleData {
            if vehicle.ownerNic == ownerNic {
                tempVehicles.push(vehicle.clone());
            }
        }
        ownedVehicles = tempVehicles.clone();
    }

    filteredVehicles.vehicles = ownedVehicles;

    DriverLicense? foundLicense = null;
    lock {

        foreach var license in licenseData {
            if license.ownerNic == ownerNic {
                foundLicense = license.clone();
            }
        }
    }
    filteredVehicles.license = foundLicense;
    return filteredVehicles.clone();
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
|};

public type DriverLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string? photoUrl;
    string ownerNic;
|};


public type VehicleClassResponse record {|
    VehicleClass[] data;
|};