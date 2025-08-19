import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;
import ballerina/os;

configurable int port = ?;

// Read environment variables
configurable string serviceURL = os:getEnv("CHOREO_MOCK_DMT_CONNECTION_SERVICEURL");
configurable string choreoApiKey = os:getEnv("CHOREO_MOCK_DMT_CONNECTION_APIKEY");

// print the consumerKey and consumerSecret

isolated service class DMTAPIClient {
    private final http:Client apiClient;

    function init() returns http:ClientError? {
        // print the service url
        log:printInfo("DMTAPIClient: Initializing", serviceURL = serviceURL, apiKey = choreoApiKey);
        self.apiClient = check new (serviceURL);
    }

    function getLicenseById(string licenseId) returns DriverLicense|error {
        log:printInfo("DMTAPIClient: Fetching license from external API", licenseId = licenseId);
        string path = string `/license/${licenseId}`;
        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
    }

    isolated function getVehicles(string? ownerNic, int skip = 0, int 'limit = 10) returns VehicleInfoResponse|error {
        log:printInfo("DMTAPIClient: Fetching vehicle info by owner", ownerNic = ownerNic);
        string path = string `/vehicle`;

        if ownerNic is string {
            string arg = string `?ownerNic=${ownerNic}`;
            path += arg;
        }

        int page = skip / 'limit;
        int pageSize = 'limit;

        path += string `?page=${page}&pageSize=${pageSize}`;

        // print ownerNic
        log:printInfo("DMTAPIClient: Fetching vehicle info by owner", path = path);

        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
    }

    isolated function getVehicleById(string vehicleId) returns VehicleInfo|error {
        log:printInfo("DMTAPIClient: Fetching vehicle info by ID", vehicleId = vehicleId);
        string path = string `/vehicles/${vehicleId}`;
        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
    }

    isolated function getVehicleByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
        log:printInfo("DMTAPIClient: Fetching vehicle info by registration number", registrationNumber = registrationNumber);
        string path = string `/vehicle/regNo/${registrationNumber}`;
        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
    }

    isolated function getDriverLicensesByOwnerNic(string ownerNic) returns DriverLicense|error {
        log:printInfo("DMTAPIClient: Fetching driver licenses by owner NIC", ownerNic = ownerNic);
        string path = string `/license/nic/${ownerNic}`;
        return self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});
    }

    isolated function getVehicleClasses() returns VehicleClass[]|error {
        log:printInfo("DMTAPIClient: Fetching vehicle classes");
        string path = string `/vehicle/types`;
        VehicleClassResponse|error response = self.apiClient->get(path, {"Choreo-API-Key": choreoApiKey});

        if response is VehicleClassResponse {
            return response.data;
        }
        return error("Failed to fetch vehicle classes");
    }
}

// This function initializes the DMTAPIClient and is used in the main GraphQL service.
public function initializeDMTClient() returns DMTAPIClient|error {
    return new ();
}

// Shared instance of the DMTAPIClient to be used across the service.
// This is initialized once and used for all requests to avoid creating multiple clients.
final DMTAPIClient sharedDMTClient = check initializeDMTClient();

# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@subgraph:Subgraph
isolated service / on new graphql:Listener(port, httpVersion = http:HTTP_1_1, host = "0.0.0.0") {
    // print the service port to the console
    public isolated function init() {
        log:printInfo("DMT service is running on port: " + port.toString());
    }

    isolated resource function get vehicle/vehicleInfoById(string vehicleId) returns VehicleInfo|error {
        return sharedDMTClient.getVehicleById(vehicleId);
    }

    isolated resource function get vehicle/vehicleInfoByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
        return sharedDMTClient.getVehicleByRegistrationNumber(registrationNumber);
    }

    // New resolver to fetch all vehicles.
    isolated resource function get vehicle/getVehicleInfos(string? ownerNic) returns VehicleInfoResponse|error {
        return sharedDMTClient.getVehicles(ownerNic);
    }

    isolated resource function get vehicle/driverLicenseById(string licenseId) returns DriverLicense|error {
        lock {
            DriverLicense[] licenses = licenseData.toArray().clone();
            foreach var license in licenses {
                if license.id == licenseId {
                    return license.clone();
                }
            }
        }
        return error("Driver license not found");
    }

    isolated resource function get vehicle/driverLicensesByOwnerId(string ownerNic) returns DriverLicense|error {
        return sharedDMTClient.getDriverLicensesByOwnerNic(ownerNic);
    }

    isolated resource function get vehicle/vehicleClasses() returns VehicleClass[]|error {
        return sharedDMTClient.getVehicleClasses();
    }

    isolated resource function get vehicle/vehicleClassById(string classId) returns VehicleClass|error {
        lock {
            VehicleClass? vehicleClass = vehicleClassData.get(classId);
            if vehicleClass is () {
                return error("Vehicle class not found");
            }
            return vehicleClass.clone();
        }
    }
}
