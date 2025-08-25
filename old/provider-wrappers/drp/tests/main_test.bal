import ballerina/test;
import ballerina/http;

// --- Mock Function for Client Initialization ---
// This annotation tells the test runner to replace the real 'initializeDRPClient'
// function from 'main.bal' with this mock version during the test run.
@test:Mock {
    moduleName: "tmp/drp_provider",
    functionName: "initializeDRPClient"
}
public function mockInitializeDRPClient() returns DRPAPIClient|error {
    // Create a mock object of the DRPAPIClient.
    DRPAPIClient mockClient = test:mock(DRPAPIClient);

    // Prepare the mock client to respond to the 'getPersonByNic' call for the SUCCESS case.
    test:prepare(mockClient).when("getPersonByNic").withArguments("199512345678").thenReturn(getMockPersonData());
    
    // Prepare the mock client to respond with an error for the NOT FOUND case.
    test:prepare(mockClient).when("getPersonByNic").withArguments("000000000000").thenReturn(error("Person not found in mock"));
    return mockClient;
}

// Helper function to return the mock data record. This uses the 'PersonData' type from 'main.bal'.
function getMockPersonData() returns PersonData {
    return {
        nic: "199512345678", fullName: "Nuwan Fernando", otherNames: "", permanentAddress: "105 Bauddhaloka Mawatha, Colombo 00400", profession: "Software Engineer", photo: "https://example.com/photo.jpg"
    };
}


// --- Test Suite ---
// The test runner will automatically start the service from 'main.bal' on port 9091.
http:Client testClient = check new ("http://localhost:9091");
@test:Config {}
public function testHealthCheck() returns error? {
    string query = "{ drp { health } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "health": "OK" } } };
    test:assertEquals(response, expected, "Response mismatch for health check");
}

@test:Config {}
public function testPersonFound() returns error? {
    string query = "{ drp { person(nic: \"199512345678\") { fullName nic civilStatus } } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "person": { "fullName": "Nuwan Fernando", "nic": "199512345678", "civilStatus": "MARRIED" } } } };
    test:assertEquals(response, expected, "Response mismatch for found person");
}

@test:Config {}
public function testPersonNotFound() returns error? {
    string query = "{ drp { person(nic: \"000000000000\") { fullName } } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "person": null } } };
    test:assertEquals(response, expected, "Response mismatch for not found person");
}

@test:Config {}
public function testCardStatusQuery() returns error? {
    string query = "{ drp { cardStatus(nic: \"199512345678\") } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "cardStatus": "ACTIVE" } } };
    test:assertEquals(response, expected, "Response mismatch for cardStatus query");
}

@test:Config {}
public function testCardStatusNotFound() returns error? {
    string query = "{ drp { cardStatus(nic: \"000000000000\") } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "cardStatus": null } } };
    test:assertEquals(response, expected, "Response mismatch for cardStatus not found");
}

@test:Config {}
public function testParentInfoQuery() returns error? {
    string query = "{ drp { parentInfo(nic: \"199512345678\") { fatherName motherName } } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "parentInfo": { "fatherName": "Father Fernando", "motherName": "Ruby de Silva" } } } };
    test:assertEquals(response, expected, "Response mismatch for parentInfo query");
}

@test:Config {}
public function testParentInfoNotFound() returns error? {
    string query = "{ drp { parentInfo(nic: \"000000000000\") { fatherName } } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "parentInfo": null } } };
    test:assertEquals(response, expected, "Response mismatch for parentInfo not found");
}

@test:Config {}
public function testLostCardInfoQuery() returns error? {
    // The mock data has no lost card info, so we expect a null response.
    string query = "{ drp { lostCardInfo(nic: \"199512345678\") { policeStation } } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "drp": { "lostCardInfo": null } } };
    test:assertEquals(response, expected, "Response mismatch for lostCardInfo query");
}