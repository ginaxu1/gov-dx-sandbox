import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/log;


@subgraph:Entity {
    key: ["id"]
}
public type User record {|
    readonly string id;
    string name;
    string dateOfBirth;
|};

isolated final table<User> key(id) userData = table [
    {id: "u-123", name: "John Doe", dateOfBirth: "1990-01-15"},
    {id: "u-456", name: "Jane Smith", dateOfBirth: "1985-05-20"}
];

@subgraph:Subgraph
isolated service / on new graphql:Listener(9091) {

    private final table<User> key(id) users;

    function init() {
        lock { 
	        self.users = userData.clone();
        }
    }

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