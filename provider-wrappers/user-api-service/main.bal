import ballerina/http;
import ballerina/log;

// A simple in-memory data store for the mock API
isolated final table<User> key(id) mockUserData = table [
    {id: "u-123", name: "John Doe", dateOfBirth: "1990-01-15"},
    {id: "u-456", name: "Jane Smith", dateOfBirth: "1985-05-20"},
    {id: "u-789", name: "Citizen Fernando", dateOfBirth: "1995-11-01"}
];

public type User record {|
    readonly string id;
    string name;
    string dateOfBirth;
|};

service /users on new http:Listener(8080) { // This service runs on port 8080

    resource function get [string id]() returns User|error {
        log:printInfo("Mock User API: Request for user", id = id);
        lock {
            User? user = mockUserData.get(id);
            if user is () {
                return error("User not found");
            }
            return user.clone();
        }
    }
}
