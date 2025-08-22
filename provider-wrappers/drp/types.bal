import ballerina/graphql.subgraph;
// This file centralizes all the data structures for the DRP service.

// This is the main entity for the subgraph.
public type PersonInfo record {|
    readonly string nic;
    string fullName;
    string otherNames;
    string permanentAddress;
    string profession;
    string contactNumber;
    string email;
    string photo;
|};

// This is the combined record for the full data set.
@subgraph:Entity {
    key: "nic"
}
public type PersonData record {|
    *PersonInfo;
|};
