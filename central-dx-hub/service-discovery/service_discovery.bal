// Jul 21 @14:55 - TODO: can't get past errors in bal build. So Choreo deployment fails. Try GO or PYthon to get past this instead of wasting time debugging Ballerina.
import ballerina/http;
import ballerina/log;
import ballerina/os;

// A map to store mocked service registrations. Mark this as 'isolated' to ensure thread-safety.
isolated map<json> mockServices = {};

// The HTTP listener for the service. Choreo will inject the PORT environment variable.
// We default to 9090 for local testing if PORT is not set.

// This is the workaround for older Ballerina versions
string portEnv = os:getEnv("PORT");
string port = portEnv == "" ? "9090" : portEnv;

listener http:Listener dxListener = new (
    check int:fromString(port)
);

service / on dxListener {

    isolated resource function get health() returns json {
        return { "status": "Ballerina Service Discovery is healthy!" };
    }

    isolated resource function post registerService(@http:Payload json payload) returns http:Created|error {
        // Use 'check' to safely extract and type-assert from json payload
        // This will propagate an error if 'name', 'url', or 'schema_path' are missing or wrong type
        string serviceName = check payload.name.ensureType();
        string serviceUrl = check payload.url.ensureType();
        string schemaPath = check payload.schema_path.ensureType();
        lock {
        mockServices[serviceName] = {
            "url": serviceUrl,
            "schema_path": schemaPath
        };
        }
        log:printInfo(string `Registered service: ${serviceName} at ${serviceUrl}`);
        // Return a 201 Created response with a message in the body
        http:Created createdResponse = {
            body: { "message": string `Service '${serviceName}' registered successfully.` }
        };

        return createdResponse;
    }
    isolated resource function get getService(string serviceName) returns json|http:NotFound {
        json? serviceInfo;
        lock {
            serviceInfo = mockServices.get(serviceName).clone();
        }

        if serviceInfo is () {
            // If not found, create and return the NotFound response.
            http:NotFound notFoundResponse = {  
                body: { "error": string `Service '${serviceName}' not found.` }
            };
            return notFoundResponse;
        }
        
        // If found, return the json value directly.
        return serviceInfo;
    }
}