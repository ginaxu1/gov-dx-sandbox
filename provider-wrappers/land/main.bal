import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

configurable int PORT = ?;

@subgraph:Subgraph
isolated service / on new graphql:Listener(PORT, httpVersion = http:HTTP_1_1, host = "0.0.0.0") {
    public isolated function init() {
        log:printInfo("Land service is running on port: " + PORT.toString());
    }

    isolated resource function get allLands() returns Land[]|error {
        lock {
            Land[][] nestedLands = from var [_, lands] in landData.entries()
                select lands;

            Land[] allLands = [];

            // Iterate through the nested arrays and add their contents to the new array
            foreach Land[] lands in nestedLands {
                foreach Land l in lands {
                    allLands.push(l);
                }
            }
            return allLands.clone();
        }
    }

    isolated resource function get landByNic(string nic) returns Land[]|error {
        lock {
            Land[]? ownedLands = landData[nic];
            return ownedLands.clone() ?: [];
        }
    }

    isolated resource function get getPersonByNic(string nic) returns PersonData? {
        PersonData|error personData;
        lock {
            Land[]? ownedLands = landData[nic];
            personData = {
                nic: nic,
                ownedLands: ownedLands.clone() ?: []
            };
        }
        if personData is error {
            log:printWarn("Land Service: Person not found or error fetching land data", nic = nic, err = personData.toString());
            return ();
        }
        return personData;
    }
}