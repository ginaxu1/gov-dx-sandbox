import ballerina/http;
import ballerina/log;

// --- Apollo Router Client ---
// An HTTP client that points to the Apollo Router service.
final http:Client apolloRouterClient = check new ("http://localhost:4000/");

// --- Gateway Proxy Service ---
// Listens on a public-facing port (e.g., 9090) and forwards
// all traffic to the internal Apollo Router.
isolated service /graphql on new http:Listener(9090) {

    // TODO: Add authentication. For now, this resource function blindly forwards any POST request.
    resource function post .(http:Request req) returns http:Response|error {
        log:printInfo("Proxying request to Apollo Router...");

        // Forward the request to the Apollo Router.
        // The Apollo Router will handle the request and return the appropriate response.
        // This is a simple proxy, so we do not modify the request in any way.
        http:Response|error response = apolloRouterClient->forward("/", req);

        // If the router is down, this will return a connection error.
        return response;
    }
}
