import ballerina/graphql;
import ballerina/log;

// Data types remain the same, but without annotations.
public type User record {|
    readonly string id;
    string name;
    string dateOfBirth;
|};

isolated final table<User> key(id) userData = table [
    {id: "u-123", name: "John Doe", dateOfBirth: "1990-01-15"},
    {id: "u-456", name: "Jane Smith", dateOfBirth: "1985-05-20"}
];

@graphql:ServiceConfig {}
isolated service / on new graphql:Listener(9091) {

    private final table<User> key(id) users;
    function init() {
        lock {
	        self.users = userData.clone();
        }
    }

    // This resource function automatically maps to the 'user' query in the schema file.
    resource function get user(string id) returns User? {
        log:printInfo("ROP Service: Looking for user", id = id);
        lock {
	        return self.users.get(id).clone();
        }
    }

    public function __resolveReference(User representation) returns User? {
        log:printInfo("ROP Service: Resolving reference for User", userId = representation.id);
        lock {
	        return self.users.get(representation.id).clone();
        }
    }
}