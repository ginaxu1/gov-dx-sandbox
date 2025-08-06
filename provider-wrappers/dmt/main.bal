import ballerina/graphql;
import ballerina/http;
import ballerina/graphql.subgraph;



# 10.5.1.1 The @subgraph:Subgraph Annotation https://ballerina.io/spec/graphql/
@subgraph:Subgraph
isolated service / on new graphql:Listener(9092, httpVersion = http:HTTP_1_1) {
    resource function get dmt/ health() returns string {
        return "OK";
    }
    resource function get dmt/ vehicleInfoById(string vehicleId) returns VehicleInfo|error {
        foreach var vehicle in vehicleData {
            if vehicle.id == vehicleId {
                return vehicle;
            }
        }
        return error("Vehicle not found");
    }

    resource function get dmt/ vehicleInfoByRegistrationNumber(string registrationNumber) returns VehicleInfo|error {
        foreach var vehicle in vehicleData {
            if vehicle.registrationNumber == registrationNumber {
                return vehicle;
            }
        }
        return error("Vehicle not found with the given registration number");
    }

    // New resolver to fetch all vehicles.
    resource function get dmt/ getVehicleInfos(string? ownerId) returns VehicleInfo[]|error {
        if ownerId is string {
            return from var vehicle in vehicleData
                   where vehicle.ownerId == ownerId
                   select vehicle;
        }
        return vehicleData.toArray();
    }

    resource function get dmt/ driverLicenseById(string licenseId) returns DriverLicense|error {
        foreach var license in licenseData {
            if license.id == licenseId {
                return license;
            }
        }
        return error("Driver license not found");
    }

    resource function get dmt/ driverLicensesByOwnerId(string ownerId) returns DriverLicense[]|error {
        return from var license in licenseData
               where license.ownerId == ownerId
               select license;
    }

    resource function get dmt/ vehicleClasses() returns VehicleClass[]|error {
        return vehicleClassData.toArray();
    }

    resource function get dmt/ vehicleClassById(string classId) returns VehicleClass|error {
        foreach var vehicleClass in vehicleClassData {
            if vehicleClass.id == classId {
                return vehicleClass;
            }
        }
        return error("Vehicle class not found");
    }
}