import ballerina/graphql;
import ballerina/graphql.subgraph;

@subgraph:Entity {
    key: "id",
    resolveReference: resolveReference
}
public type User record {|
    readonly string id;
    string name?;
    boolean hasDriversLicense;
    DriversLicense? driversLicense;
|};

public type DriversLicense record {|
    readonly string id;
    string licenseNumber;
    string expiryDate;
|};

isolated function resolveReference(map<anydata> representation) returns User|error {
    string id = <string>representation["id"];
    if id == "u-123" {
        return {
            id: id,
            hasDriversLicense: true,
            driversLicense: {
                id: "dl-987",
                licenseNumber: "B1234567",
                expiryDate: "2028-01-15"
            }
        };
    }
    return error("User not found");
}
# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@subgraph:Subgraph
isolated service / on new graphql:Listener(9090) {
    resource function get health() returns string {
        return "OK";
    }
}