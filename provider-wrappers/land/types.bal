import ballerina/graphql.subgraph;

public type Land record {
    string landRegistry;
    string location;
    string landRegistryDistrict;
    string address;
    string officeMobile?;
    string telephoneNumber;
    string officeFax?;
    string officeEmail?;
    string[] dsListCovered;
};

@subgraph:Entity {
    key: "nic"
}
public type PersonData record {|
    readonly string nic;
    Land[] ownedLands;
|};

isolated function resolvePersonData(subgraph:Representation representation) returns PersonData|error {
    string nic = check representation["nic"].ensureType();
    Land[]? ownedLands;
    lock {
        ownedLands = landData[nic].clone();
    }

    return {
        nic: nic,
        ownedLands: ownedLands ?: []
    };
}