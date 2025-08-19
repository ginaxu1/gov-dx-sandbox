import ballerina/graphql.subgraph;
import ballerina/log;

@subgraph:Entity {
    key: "id ownerNic",
    resolveReference: resolveVehicleData
}
public type VehicleInfo record {|
    readonly string id;
    readonly string ownerNic;
    PersonData? owner;
|};

isolated function resolveVehicleData(subgraph:Representation representation) returns VehicleInfo|error? {
    string id = check representation["id"].ensureType();
    string ownerNic = check representation["ownerNic"].ensureType();

    // log
    log:printInfo("Resolving vehicle data for id: " + id + ", ownerNic: " + ownerNic);

    VehicleInfo vehicle = {
        id: id,
        ownerNic: ownerNic,
        owner: null
    };

    PersonData|error ownerData = sharedDRPClient.getPersonByNic(ownerNic);

    if ownerData is PersonData {
        vehicle.owner = ownerData;
    }

    return vehicle;
}
