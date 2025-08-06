import ballerina/graphql;
import ballerina/http;
import ballerina/log;

// This gateway implements the "namespace wrapping" or "schema stitching" pattern.
// It exposes a single GraphQL endpoint that proxies requests to the underlying
// DRP and DMT subgraphs. It does NOT use GraphQL Federation.

// --- Subgraph Clients ---
// The gateway will communicate with each subgraph using these clients.
// The URLs point to the individual GraphQL endpoints of our microservices.
final graphql:Client drpClient = check new ("http://localhost:9091/");
final graphql:Client dmtClient = check new ("http://localhost:9092/");

//  --- Gateway Service  --- 
// Listens on port 9090 and acts as the single entry point for all client applications.
isolated service /graphql on new http:Listener(9090) {

    // This resource function handles GraphQL queries sent via HTTP POST.
    resource function post .(http:Request req) returns json|error {
        // Decode the incoming GraphQL JSON payload.
        // Payload should be in the format: { "query": "...", "variables": {...} }
        map<json> payload = check req.getJsonPayload().ensureType();
        string query = check payload.query.ensureType();
        map<json> variables = payload.hasKey("variables") ? check payload.variables.ensureType() : {};

        // Start two "futures" to run the network calls in parallel.
        future<json|error> drpFuture = start callService("drp", query, variables, drpClient);
        future<json|error> dmtFuture = start callService("dmt", query, variables, dmtClient);

        // Wait for both calls to complete.
        json|error drpResult = wait drpFuture;
        json|error dmtResult = wait dmtFuture;

        // Merge the results from both services.
        map<json> finalData = {};
        json[] finalErrors = [];
        processResult(drpResult, finalData, finalErrors);
        processResult(dmtResult, finalData, finalErrors);

        // Construct the final response payload.
        map<json> finalResponse = {};
        if finalData.length() > 0 {
            finalResponse["data"] = finalData;
        }
        if finalErrors.length() > 0 {
            finalResponse["errors"] = finalErrors;
        }
        return finalResponse;
    }
}

// --- Helper Functions ---
// Calls a downstream GraphQL service only if its corresponding field is mentioned in the query.
function callService(string fieldName, string query, map<json> variables, graphql:Client gqlClient) returns json|error {
    // Only execute the query against the service if the query string contains its root field.
    // This prevents unnecessary network calls.
    if query.includes(fieldName) {
        log:printInfo(string `Proxying request to ${fieldName} service`);
        return check gqlClient->execute(query, variables);
    }
    // If the field is not in the query, return an empty map.
    return {};
}

// Merges the data and errors from a single service call into the final response maps.
function processResult(json|error result, map<json> finalData, json[] finalErrors) {
    if result is map<json> {
        // Merge the data object if it exists.
        if result.hasKey("data") && result.data is map<json> {
            // This approach iterates over the keys and is more robust.
            var dataField = result.data;
            if dataField is map<json> {
                foreach string key in dataField.keys() {
                    finalData[key] = dataField[key];
                }
            }
        }
        // Append all errors from the service's error array if it exists.
        if result.hasKey("errors") {
            var errorsArr = result.errors;
            if errorsArr is json[] {
                finalErrors.push(...errorsArr);
            }
        }
    } else if result is error {
        log:printError("Error from downstream service", 'error = result);
        finalErrors.push({ message: result.message(), path: [result.toString()] });
    }
}