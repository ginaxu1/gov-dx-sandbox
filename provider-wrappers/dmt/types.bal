import ballerina/graphql.subgraph;

public type VehicleClass record {|
    readonly string id;
    string className;
|};

@subgraph:Entity {
    key: "nic",
    resolveReference: resolveVehicleInfo
}
public type PersonData record {|
    readonly string nic;
    VehicleInfo[] vehicles;
|};

isolated function resolveVehicleInfo(subgraph:Representation representation) returns PersonData|error? {
    string ownerSludi = check representation["nic"].ensureType();

    lock {

        PersonData filteredVehicles = {nic: check representation["nic"].ensureType(), vehicles: []};
        foreach var vehicle in vehicleData {
            if vehicle.ownerId == ownerSludi {
                filteredVehicles.vehicles.push(vehicle);
            }
        }

        return filteredVehicles.clone();
    }
}

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
