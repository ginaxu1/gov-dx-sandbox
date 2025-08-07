import ballerina/graphql;
import ballerina/http;
import ballerina/graphql.subgraph;
import ballerina/log;

// read the port from the environment variable
final int servicePort = 9090;



# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@subgraph:Subgraph
isolated service / on new graphql:Listener(servicePort, httpVersion = http:HTTP_1_1) {
    // print the service port to the console
    public isolated function init() {
        log:printInfo("DMT service is running on port: " + servicePort.toString());
    }

    isolated resource function get  vehicleInfoById(string vehicleId) returns VehicleInfo|error {
        foreach var vehicle in vehicleData {
            if vehicle.id == vehicleId {
                return vehicle;
            }
        }
        return error("Vehicle not found");
    }

    isolated resource function get  vehicleInfoByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
        foreach var vehicle in vehicleData {
            if vehicle.registrationNumber == registrationNumber {
                return vehicle;
            }
        }
        return error("Vehicle not found with the given registration number");
    }

    // New resolver to fetch all vehicles.
    isolated resource function get getVehicleInfos(string? ownerId) returns VehicleInfo[]|error {
        lock {
            if ownerId is string {
                return from var vehicle in vehicleData
                       where vehicle.ownerId == ownerId
                       select vehicle;
            }
        }

        return vehicleData.toArray();
        }

    isolated resource function get driverLicenseById(string licenseId) returns DriverLicense|error {
        lock {            
            foreach var license in licenseData {
                if license.id == licenseId {
                    return license;
                }
            }
        }
        return error("Driver license not found");
    }

    isolated resource function get driverLicensesByOwnerId(string ownerId) returns DriverLicense[]|error {
        return from var license in licenseData
               where license.ownerId == ownerId
               select license;
    }

    isolated resource function get vehicleClasses() returns VehicleClass[]|error {
        return vehicleClassData.toArray();
    }

    isolated resource function get vehicleClassById(string classId) returns VehicleClass|error {
        foreach var vehicleClass in vehicleClassData {
            if vehicleClass.id == classId {
                return vehicleClass;
            }
        }
        return error("Vehicle class not found");
    }
}