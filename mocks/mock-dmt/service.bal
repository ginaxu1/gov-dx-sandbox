import mock_dmt.store;

import ballerina/http;
import ballerina/persist;

configurable int port = ?;

final store:Client sClient = check new ();

# A service representing a network-accessible API
# bound to port `9093`.
service / on new http:Listener(port) {

    resource function get license/owner(int page = 1, int pageSize = 10) returns json|error {
        stream<store:OwnerInfo, persist:Error?> resultStream = sClient->/ownerinfos();

        // Calculate how many items to skip
        int skipCount = (page - 1) * pageSize;

        // Manually paginate the stream
        store:OwnerInfo[] paginatedOwners = [];
        int count = 0;
        persist:Error? err = ();

        while true {
            var next = resultStream.next();
            if next is record {|store:OwnerInfo value;|} {
                if count >= skipCount && paginatedOwners.length() < pageSize {
                    paginatedOwners.push(next.value);
                }
                count += 1;
                if paginatedOwners.length() == pageSize {
                    break;
                }
            } else {
                err = next;
                break;
            }
        }

        if err is persist:Error {
            return err;
        }

        return {
            data: paginatedOwners,
            pagination: {
                page: page,
                pageSize: pageSize,
                total: count
            }
        };
    }

    resource function post license/owner(store:OwnerInfoInsert ownerInfo) returns json|error {

        string[]| error? result = sClient->/ownerinfos.post([ownerInfo]);

        if result is string[] {
            return {
                id: result[0]
            };
        }
        return error("Failed to create owner info");
    }

    resource function get license/issuer() returns json|error {
        stream<store:IssuerInfo, persist:Error?> resultStream = sClient->/issuerinfos();

        // Collect all results
        store:IssuerInfo[] issuerInfos = [];
        persist:Error? err = ();

        while true {
            var next = resultStream.next();
            if next is record {|store:IssuerInfo value;|} {
                issuerInfos.push(next.value);
            } else {
                err = next;
                break;
            }
        }

        if err is persist:Error {
            return err;
        }

        return {
            data: issuerInfos
        };
    }

    resource function post license/issuer(store:IssuerInfoInsert issuerInfo) returns json|error {

        string[]| error? result = sClient->/issuerinfos.post([issuerInfo]);

        if result is string[] {
            return {
                id: result[0]
            };
        }
        return error("Failed to create issuer info");
    }

    resource function post license(store:DrivingLicenseInsert license) returns json|error {

        string[]| error? result = sClient->/drivinglicenses.post([license]);

        if result is string[] {
            return {
                id: result[0]
            };
        }
        return error("Failed to create driving license");
    }

    resource function get license(int page = 1, int pageSize = 10) returns json|error {
        stream<store:DrivingLicenseWithRelations, persist:Error?> resultStream = sClient->/drivinglicenses();

        // Calculate how many items to skip
        int skipCount = (page - 1) * pageSize;

        // Manually paginate the stream
        store:DrivingLicenseWithRelations[] paginatedLicenses = [];
        int count = 0;
        persist:Error? err = ();

        while true {
            var next = resultStream.next();
            if next is record {|store:DrivingLicenseWithRelations value;|} {
                if count >= skipCount && paginatedLicenses.length() < pageSize {
                    paginatedLicenses.push(next.value);
                }
                count += 1;
                if paginatedLicenses.length() == pageSize {
                    break;
                }
            } else {
                err = next;
                break;
            }
        }

        if err is persist:Error {
            return err;
        }

        return {
            data: paginatedLicenses,
            pagination: {
                page: page,
                pageSize: pageSize,
                total: count
            }
        };
    }

    isolated resource function get licenses/[string id]() returns DrivingLicense|http:NotFound {
        lock {
            if (!drivingLicenses.hasKey(id)) {
                return http:NOT_FOUND;
            }
            DrivingLicense license = drivingLicenses.get(id);
            return license.clone();
        }
    }
}
