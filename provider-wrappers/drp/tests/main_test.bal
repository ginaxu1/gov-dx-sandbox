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
    
    // Prepare the mock client to respond to the 'getPersonByNic' call for the NOT FOUND case.
    test:prepare(mockClient).when("getPersonByNic").withArguments("000000000000").thenReturn(error("Person not found in mock"));

    return mockClient;
}

// Helper function to return the mock data record.
// This uses the 'PersonData' type from 'main.bal'.
function getMockPersonData() returns PersonData {
    return {
        nic: "199512345678", fullName: "Nuwan Fernando", surname: "Fernando", otherNames: "Nuwan", gender: MALE, dateOfBirth: "1995-12-01", placeOfBirth: "Colombo", permanentAddress: "105 Bauddhaloka Mawatha, Colombo 00400", profession: "Software Engineer", civilStatus: MARRIED, contactNumber: "+94771234567", email: "nuwan@opensource.lk", photo: "https://example.com/photo.jpg",
        cardInfo: { cardNumber: "199512345678", issueDate: "2018-01-02", expiryDate: "2028-01-01", cardStatus: ACTIVE },
        lostCardReplacementInfo: (),
        citizenshipInfo: { citizenshipType: DESCENT, certificateNumber: "A12345", issueDate: "1995-12-02" },
        parentInfo: { fatherName: "Father Fernando", motherName: "Ruby de Silva", fatherNic: "196618234567", motherNic: "196817654321" }
    };
}


// --- Test Suite ---
// This test suite uses the mock client to simulate the behavior of the DRP service.
http:Client testClient = check new ("http://localhost:9091");

@test:Config {}
public function testPersonFound() returns error? {
    string query = "{ person(nic: \"199512345678\") { fullName nic civilStatus } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "person": { "fullName": "Nuwan Fernando", "nic": "199512345678", "civilStatus": "MARRIED" } } };
    test:assertEquals(response, expected, "Response mismatch for found person");
}

@test:Config {}
public function testPersonNotFound() returns error? {
    string query = "{ person(nic: \"000000000000\") { fullName } }";
    json payload = { "query": query };
    json response = check testClient->post("/", payload);
    json expected = { "data": { "person": null } };
    test:assertEquals(response, expected, "Response mismatch for not found person");
}
