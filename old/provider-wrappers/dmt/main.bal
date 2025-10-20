import ballerina/graphql;
import ballerina/http;
import ballerina/log;

configurable int port = ?;

// // Shared instance of the DMTAPIClient to be used across the service.
// // This is initialized once and used for all requests to avoid creating multiple clients.
// # 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
// @subgraph:Subgraph
// isolated service / on new graphql:Listener(port, httpVersion = http:HTTP_1_1, host = "0.0.0.0") {
//     // print the service port to the console
//     public isolated function init() {
//         log:printInfo("DMT service is running on port: " + port.toString());
//     }

//     isolated resource function get vehicle/ vehicleInfoById(string vehicleId) returns VehicleInfo|error {
//         return sharedDMTClient.getVehicleById(vehicleId);
//     }

//     isolated resource function get vehicle/ vehicleInfoByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
//         return sharedDMTClient.getVehicleByRegistrationNumber(registrationNumber);
//     }

//     // New resolver to fetch all vehicles.
//     isolated resource function get vehicle/ getVehicleInfos(string? ownerNic) returns VehicleInfoResponse|error {
//         return sharedDMTClient.getVehicles(ownerNic);
//     }


//     isolated resource function get vehicle/ driverLicenseById(string licenseId) returns DriverLicense|error {
//         lock {
//             DriverLicense[] licenses = licenseData.toArray().clone();
//             foreach var license in licenses {
//                 if license.id == licenseId {
//                     return license.clone();
//                 }
//             }
//         }
//         return error("Driver license not found");
//     }

//     isolated resource function get vehicle/ driverLicensesByOwnerId(string ownerNic) returns DriverLicense|error {
//         return sharedDMTClient.getDriverLicensesByOwnerNic(ownerNic);
//     }

//     isolated resource function get vehicle/ vehicleClasses() returns VehicleClass[]|error {
//         return sharedDMTClient.getVehicleClasses();
//     }

//     isolated resource function get vehicle/ vehicleClassById(string classId) returns VehicleClass|error {
//         lock {
//             VehicleClass? vehicleClass = vehicleClassData.get(classId);
//             if vehicleClass is () {
//                 return error("Vehicle class not found");
//             }
//             return vehicleClass.clone();
//         }
//     }
// }
@graphql:ServiceConfig {
    graphiql: {
        enabled: true
    }
}
service / on new graphql:Listener(port, httpVersion = http:HTTP_1_1, host = "0.0.0.0") {
    // print the service port to the console
    public isolated function init() {
        log:printInfo("DMT service is running on port: " + port.toString());
    }

    resource function get vehicles(string? ownerNic) returns VehicleInfo[]|error {
        lock {
            VehicleInfo[] allVehicles = vehicleData.toArray().clone();
            if ownerNic is string {
                allVehicles = allVehicles.filter(vehicle => vehicle.ownerNic == ownerNic);
            }
            return allVehicles.clone();
        }
    }
}
            