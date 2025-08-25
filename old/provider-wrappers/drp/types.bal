import ballerina/graphql.subgraph;
import ballerina/graphql;
// This file centralizes all the data structures for the DRP service.

// This is the main entity for the subgraph.
// This is the combined record for the full data set.

@subgraph:Entity {
    key: "nic"
}
public type PersonData record {|
    @graphql:ID readonly string nic;
    string fullName;
    string otherNames;
    string permanentAddress;
    string profession;
    string photo;
    anydata...;
|};
