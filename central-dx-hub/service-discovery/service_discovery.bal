import ballerina/http;
import ballerina/log;
import ballerina/runtime; 

// A map to store mocked service registrations.
// In a real system, this would be backed by a persistent database.
// Structure: { "service_name": { "url": "...", "schema_path": "..." } }
map<json> mockServices = {};

// The HTTP listener for the service. Choreo will inject the PORT environment variable.
listener http:Listener dxListener = new (int:fromString(runtime:getEnv("PORT") ?: "9090"));

service / on dxListener { // Use the named listener here

    // Health check endpoint.
    // Returns a simple status to indicate the service is running.
    resource function get health() returns json {
        return { "status": "Ballerina Service Discovery is healthy!" };
    }

    // Endpoint to register a new service.
    // Expects a JSON payload with 'name', 'url', and 'schema_path'.
    resource function post registerService(@http:Payload json payload) returns http:Response|http:BadRequest {
        string? serviceName = payload.name.ensureType(string); // Safer type assertion
        string? serviceUrl = payload.url.ensureType(string);
        string? schemaPath = payload.schema_path.ensureType(string);

        if serviceName is () || serviceUrl is () || schemaPath is () {
            // Return http:BadRequest directly with the error message
            return new http:BadRequest(
                body = { "error": "Missing 'name', 'url', or 'schema_path' in request" }
            );
        }

        mockServices[serviceName] = {
            "url": serviceUrl,
            "schema_path": schemaPath
        };
        log:printInfo(string `Registered service: ${serviceName} at ${serviceUrl}`);
        
        // Return http:Response directly
        http:Response res = new;
        res.statusCode = http:CREATED; // Use the constant directly for status code
        res.setJsonPayload({ "message": string `Service '${serviceName}' registered successfully.` });
        return res;
    }

    // Endpoint to retrieve details of a registered service by name.
    resource function get getService(@http:Path { value: "serviceName" } string serviceName) returns json|http:NotFound {
        json? serviceInfo = mockServices[serviceName];
        if serviceInfo is json {
            return serviceInfo;
        } else {
            // Return http:NotFound directly with the error message
            return new http:NotFound(
                body = { "error": string `Service '${serviceName}' not found.` }
            );
        }
    }
}