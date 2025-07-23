import ballerina/graphql;

// This record directly maps to the 'DriverLicense' type in our schema.
// Ballerina automatically links them by name.
type DriverLicense record {|
    string id;
    string name;
    string licenseNumber;
    string issueDate;
    string expiryDate;
    string? photoUrl; // nullable field, can be omitted
|};

// This creates an HTTP service that listens for GraphQL requests on port 9090.
service / on new graphql:Listener(9090) {

    // This resource method with 'get' accessor is required by Ballerina GraphQL services.
    resource function get health() returns string {
        return "OK";
        }
    // This remote function is the 'resolver' for the 'driverLicenseById' query. Ballerina will automatically link it to the GraphQL schema.
    // It takes an 'id' argument and returns a 'DriverLicense' record or null if not found.
        remote function driverLicenseById(string id) returns DriverLicense? {
        // For sandbox, return static mocked data.
        if id == "197419202757" {
            return {
                id: "197419202757",
                name: "Test User",
                licenseNumber: "AAA0001",
                issueDate: "2024-08-01",
                expiryDate: "2029-08-01",
                photoUrl: "http://example.com/photo.jpg"
            };
        }
        return (); 
    }
}