import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/log;

// Represents the 'DriverLicense' type in the GraphQL schema.
public type DriverLicense record {|
    readonly string id;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string? photoUrl;
    string ownerId; // Internal link to a User
|};

// Defines the User as a federated entity, specifying its key.
@subgraph:Entity {
    key: ["id"],
    resolveReference: resolveReference
}
public type User record {|
    readonly string id;
    DriverLicense? driversLicense;
    boolean hasDriversLicense;
|};

// Mock data table for driver licenses.
isolated final table<DriverLicense> key(id) licenseData = table [
    {id: "dl-abc", licenseNumber: "D12345678", issueDate: "2020-10-10", expiryDate: "2025-10-09", ownerId: "u-123", photoUrl: "http://example.com/photo1.jpg"},
    {id: "dl-def", licenseNumber: "D87654321", issueDate: "2022-01-01", expiryDate: "2027-12-31", ownerId: "u-456", photoUrl: "http://example.com/photo2.jpg"}
];

isolated function resolveReference(map<anydata> representation) returns User|error {
    string id = <string>representation["id"];
    log:printInfo("DMV Service: Resolving reference for User", userId = id);

    DriverLicense? 'license = ();
    boolean hasLicense = false;

    table<(DriverLicense & readonly)> result = table [];
    lock {
        result = from var dl in licenseData
            where dl.ownerId == id
            select dl.cloneReadOnly();
    }

    if result.length() > 0 {
        'license = result.toArray()[0];
        hasLicense = true;
    }

    return {
        id: id,
        driversLicense: 'license,
        hasDriversLicense: hasLicense
    };
}



# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@graphql:ServiceConfig {
    cors: {
        allowOrigins: ["https://studio.apollographql.com"],
        allowCredentials: false,
        allowMethods: ["GET", "POST", "OPTIONS"],
        allowHeaders: ["CORELATION_ID"],
        exposeHeaders: ["X-CUSTOM-HEADER"],
        maxAge: 84900
    }
}
isolated service / on new graphql:Listener(9092) {
    resource function get health() returns string {
        return "OK";
    }

    resource function get driverLicenses() returns DriverLicense[]|error {
        log:printInfo("DMV Service: Fetching all driver licenses");
        lock {
	        return licenseData.toArray().clone();
        }
    }
}