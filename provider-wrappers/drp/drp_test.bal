import ballerina/test;
import ballerina/http;
import ballerina/log;
import ballerina/graphql;

// Test configuration to start the main GraphQL service on a different port for testing.

// Define the GraphQL service from your main drp.bal file.
// This allows the test suite to start it.
isolated service / on new graphql:Listener(9092) { // Running on a new port 9092 for tests
    resource function get person(string nic) returns PersonData? {
        log:printInfo("Test DRP Service: Looking for person via API Client", nic = nic);
        PersonData|error personData = sharedDRPClient.getPersonByNic(nic);
        if personData is error {
            log:printWarn("Test DRP Service: Person not found or error fetching person", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }

    resource function get health() returns string {
        log:printInfo("Test DRP Service: Health check requested.");
        return "OK";
    }
}

// Global HTTP client to interact with the test service.
http:Client testClient = check new ("http://localhost:9092");

// --- Test Functions ---

// Test case for successfully finding a person.
@test:Config {}
function testPersonFound() returns error? {
    // 1. Define the GraphQL query
    string query = "{ person(nic: \"199512345678\") { fullName nic civilStatus } }";
    json payload = { "query": query };

    // 2. Send the request to the test GraphQL service
    json response = check testClient->post("/", payload);

    // 3. Define the expected response
    json expectedResponse = {
        "data": {
            "person": {
                "fullName": "Nuwan Fernando",
                "nic": "199512345678",
                "civilStatus": "MARRIED"
            }
        }
    };

    // 4. Assert that the actual response matches the expected response
    test:assertEquals(response, expectedResponse, "Response mismatch for found person");
}

// Test case for when a person is not found.
@test:Config {}
function testPersonNotFound() returns error? {
    // 1. Define the GraphQL query for a non-existent NIC
    string query = "{ person(nic: \"000000000000\") { fullName } }";
    json payload = { "query": query };

    // 2. Send the request
    json response = check testClient->post("/", payload);

    // 3. Define the expected null response
    json expectedResponse = {
        "data": {
            "person": null
        }
    };

    // 4. Assert that the service correctly returns null for the person field
    test:assertEquals(response, expectedResponse, "Response mismatch for not found person");
}