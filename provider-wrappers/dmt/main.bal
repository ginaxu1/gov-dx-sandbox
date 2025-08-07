import ballerina/graphql;
import ballerina/graphql.subgraph;
import ballerina/http;
import ballerina/log;

const servicePort = 9090;

# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@subgraph:Subgraph
isolated service / on new graphql:Listener(servicePort, httpVersion = http:HTTP_1_1) {
    // print the service port to the console
    public isolated function init() {
        log:printInfo("DMT service is running on port: " + servicePort.toString());
    }

    isolated resource function get vehicleInfoById(string vehicleId) returns VehicleInfo|error {
        lock {
            foreach var vehicle in vehicleData {
                if vehicle.id == vehicleId {
                    return vehicle.clone();
                }
            }
        }
        return error("Vehicle not found");
    }

    isolated resource function get vehicleInfoByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
        lock {
            foreach var vehicle in vehicleData {
                if vehicle.registrationNumber == registrationNumber {
                    return vehicle.clone();
                }
            }
        }
        return error("Vehicle not found with the given registration number");
    }

    // New resolver to fetch all vehicles.
    isolated resource function get getVehicleInfos(string? ownerId) returns VehicleInfo[]|error {
        lock {
            if ownerId is string {
                VehicleInfo[] filteredVehicles = [];
                foreach var vehicle in vehicleData {
                    if vehicle.ownerId == ownerId {
                        filteredVehicles.push(vehicle.clone());
                    }
                }
                return filteredVehicles.clone();
            }
            VehicleInfo[] allVehicles = vehicleData.toArray().clone();
            return allVehicles.clone();
        }
    }

    isolated resource function get driverLicenseById(string licenseId) returns DriverLicense|error {
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

    isolated resource function get driverLicensesByOwnerId(string ownerId) returns DriverLicense[]|error {
        lock {
            DriverLicense[] selectedLicenses = [];
            foreach var license in licenseData {
                if license.ownerId == ownerId {
                    selectedLicenses.push(license.clone());
                }
            }
            return selectedLicenses.clone();
        }
    }

    isolated resource function get vehicleClasses() returns VehicleClass[]|error {
        lock {
            return vehicleClassData.toArray().clone();
        }
    }

    isolated resource function get vehicleClassById(string classId) returns VehicleClass|error {
        lock {
            VehicleClass? vehicleClass = vehicleClassData.get(classId);
            if vehicleClass is () {
                return error("Vehicle class not found");
            }
            return vehicleClass.clone();
        }
    }
}
