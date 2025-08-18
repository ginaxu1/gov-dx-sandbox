// common/graphql/directives.bal
// This file defines common GraphQL types and custom directives
// that can be used across your Ballerina GraphQL subgraphs.

import ballerina/graphql;

// Define the enum for the classification levels, consistent with Go service
public enum Classification {
    ALLOW,
    ALLOW_PROVIDER_CONSENT,
    ALLOW_CITIZEN_CONSENT,
    ALLOW_CONSENT, // Generic consent, handled by context in Go service
    DENIED
}

// Define a Permission record type if needed within Ballerina for conceptual clarity
// (Though directly using Classification enum in directives is more common)
public type Permission record {
    Classification classification;
};

// Define a custom GraphQL directive for policy checking.
// This directive can be applied to fields in your GraphQL schema.
// It tells the Apollo Router that this field requires a policy evaluation.
// The 'classification' argument serves as an initial hint or the default policy for the field.
@graphql:directive {
    name: "policyCheck",
    locations: [graphql:FIELD_DEFINITION] // This directive can be applied to fields
}
public annotation @PolicyCheckDirective {
    Classification classification; // The classification level for this field
    // You could add other parameters if needed, e.g., 'resourceName', 'scope'
} on graphql:FIELD_DEFINITION;


// Example of how these might be used in a Ballerina subgraph (e.g., in dmt/main.bal's schema definition)
/*
// Assuming this is part of your dmt/main.bal or a schema definition file
type VehicleInfo record {|
    readonly string id;
    string make;
    string model;
    int yearOfManufacture;
    string ownerNic;

    // Apply the @PolicyCheckDirective to fields requiring policy evaluation
    string engineNumber @PolicyCheckDirective { classification: Classification.ALLOW_PROVIDER_CONSENT };
    string conditionAndNotes @PolicyCheckDirective { classification: Classification.ALLOW_PROVIDER_CONSENT };
    string registrationNumber @PolicyCheckDirective { classification: Classification.ALLOW };
    VehicleClass vehicleClass;
|};

type DriverLicense record {|
    readonly string id;
    // This field needs provider consent according to the database policy
    string licenseNumber @PolicyCheckDirective { classification: Classification.ALLOW_PROVIDER_CONSENT };
    string issueDate;
    string expiryDate;
    string? photoUrl;
    string ownerNic;
|};

// And in your drp subgraph's schema:
type PersonData record {|
    readonly string nic;
    // This field needs citizen consent
    string photo @PolicyCheckDirective { classification: Classification.ALLOW_CITIZEN_CONSENT };
    // ... other fields
|};
*/