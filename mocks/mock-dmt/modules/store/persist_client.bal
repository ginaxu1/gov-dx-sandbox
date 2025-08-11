// AUTO-GENERATED FILE. DO NOT MODIFY.

// This file is an auto-generated file by Ballerina persistence layer for model.
// It should not be modified by hand.

import ballerina/jballerina.java;
import ballerina/persist;
import ballerina/sql;
import ballerinax/persist.sql as psql;
import ballerinax/postgresql;
import ballerinax/postgresql.driver as _;

const DRIVING_LICENSE = "drivinglicenses";
const VEHICLE_CLASS = "vehicleclasses";
const VEHICLE_INFO = "vehicleinfos";
const OWNER_INFO = "ownerinfos";
const ISSUER_INFO = "issuerinfos";
const VEHICLE_PERMISSION = "vehiclepermissions";

public isolated client class Client {
    *persist:AbstractPersistClient;

    private final postgresql:Client dbClient;

    private final map<psql:SQLClient> persistClients;

    private final record {|psql:SQLMetadata...;|} metadata = {
        [DRIVING_LICENSE]: {
            entityName: "DrivingLicense",
            tableName: "DrivingLicense",
            fieldMetadata: {
                id: {columnName: "id"},
                licenseNumber: {columnName: "licenseNumber"},
                issueDate: {columnName: "issueDate"},
                expiryDate: {columnName: "expiryDate"},
                frontImageUrl: {columnName: "frontImageUrl"},
                backImageUrl: {columnName: "backImageUrl"},
                ownerinfoOwnerNic: {columnName: "ownerinfoOwnerNic"},
                issuerinfoId: {columnName: "issuerinfoId"},
                "permissions[].id": {relation: {entityName: "permissions", refField: "id"}},
                "permissions[].vehicleType": {relation: {entityName: "permissions", refField: "vehicleType"}},
                "permissions[].issueDate": {relation: {entityName: "permissions", refField: "issueDate"}},
                "permissions[].expiryDate": {relation: {entityName: "permissions", refField: "expiryDate"}},
                "permissions[].drivinglicenseId": {relation: {entityName: "permissions", refField: "drivinglicenseId"}},
                "ownerInfo.ownerNic": {relation: {entityName: "ownerInfo", refField: "ownerNic"}},
                "ownerInfo.name": {relation: {entityName: "ownerInfo", refField: "name"}},
                "ownerInfo.address": {relation: {entityName: "ownerInfo", refField: "address"}},
                "ownerInfo.birthDate": {relation: {entityName: "ownerInfo", refField: "birthDate"}},
                "ownerInfo.signatureUrl": {relation: {entityName: "ownerInfo", refField: "signatureUrl"}},
                "ownerInfo.bloodGroup": {relation: {entityName: "ownerInfo", refField: "bloodGroup"}},
                "issuerInfo.id": {relation: {entityName: "issuerInfo", refField: "id"}},
                "issuerInfo.name": {relation: {entityName: "issuerInfo", refField: "name"}},
                "issuerInfo.issuingAuthority": {relation: {entityName: "issuerInfo", refField: "issuingAuthority"}},
                "issuerInfo.signatureUrl": {relation: {entityName: "issuerInfo", refField: "signatureUrl"}}
            },
            keyFields: ["id"],
            joinMetadata: {
                permissions: {entity: VehiclePermission, fieldName: "permissions", refTable: "VehiclePermission", refColumns: ["drivinglicenseId"], joinColumns: ["id"], 'type: psql:MANY_TO_ONE},
                ownerInfo: {entity: OwnerInfo, fieldName: "ownerInfo", refTable: "OwnerInfo", refColumns: ["ownerNic"], joinColumns: ["ownerinfoOwnerNic"], 'type: psql:ONE_TO_ONE},
                issuerInfo: {entity: IssuerInfo, fieldName: "issuerInfo", refTable: "IssuerInfo", refColumns: ["id"], joinColumns: ["issuerinfoId"], 'type: psql:ONE_TO_MANY}
            }
        },
        [VEHICLE_CLASS]: {
            entityName: "VehicleClass",
            tableName: "VehicleClass",
            fieldMetadata: {
                id: {columnName: "id"},
                className: {columnName: "className"},
                "vehicles[].id": {relation: {entityName: "vehicles", refField: "id"}},
                "vehicles[].make": {relation: {entityName: "vehicles", refField: "make"}},
                "vehicles[].model": {relation: {entityName: "vehicles", refField: "model"}},
                "vehicles[].yearOfManufacture": {relation: {entityName: "vehicles", refField: "yearOfManufacture"}},
                "vehicles[].ownerNic": {relation: {entityName: "vehicles", refField: "ownerNic"}},
                "vehicles[].engineNumber": {relation: {entityName: "vehicles", refField: "engineNumber"}},
                "vehicles[].conditionAndNotes": {relation: {entityName: "vehicles", refField: "conditionAndNotes"}},
                "vehicles[].registrationNumber": {relation: {entityName: "vehicles", refField: "registrationNumber"}},
                "vehicles[].vehicleclassId": {relation: {entityName: "vehicles", refField: "vehicleclassId"}}
            },
            keyFields: ["id"],
            joinMetadata: {vehicles: {entity: VehicleInfo, fieldName: "vehicles", refTable: "VehicleInfo", refColumns: ["vehicleclassId"], joinColumns: ["id"], 'type: psql:MANY_TO_ONE}}
        },
        [VEHICLE_INFO]: {
            entityName: "VehicleInfo",
            tableName: "VehicleInfo",
            fieldMetadata: {
                id: {columnName: "id"},
                make: {columnName: "make"},
                model: {columnName: "model"},
                yearOfManufacture: {columnName: "yearOfManufacture"},
                ownerNic: {columnName: "ownerNic"},
                engineNumber: {columnName: "engineNumber"},
                conditionAndNotes: {columnName: "conditionAndNotes"},
                registrationNumber: {columnName: "registrationNumber"},
                vehicleclassId: {columnName: "vehicleclassId"},
                "vehicleClass.id": {relation: {entityName: "vehicleClass", refField: "id"}},
                "vehicleClass.className": {relation: {entityName: "vehicleClass", refField: "className"}}
            },
            keyFields: ["id"],
            joinMetadata: {vehicleClass: {entity: VehicleClass, fieldName: "vehicleClass", refTable: "VehicleClass", refColumns: ["id"], joinColumns: ["vehicleclassId"], 'type: psql:ONE_TO_MANY}}
        },
        [OWNER_INFO]: {
            entityName: "OwnerInfo",
            tableName: "OwnerInfo",
            fieldMetadata: {
                ownerNic: {columnName: "ownerNic"},
                name: {columnName: "name"},
                address: {columnName: "address"},
                birthDate: {columnName: "birthDate"},
                signatureUrl: {columnName: "signatureUrl"},
                bloodGroup: {columnName: "bloodGroup"},
                "drivingLicense.id": {relation: {entityName: "drivingLicense", refField: "id"}},
                "drivingLicense.licenseNumber": {relation: {entityName: "drivingLicense", refField: "licenseNumber"}},
                "drivingLicense.issueDate": {relation: {entityName: "drivingLicense", refField: "issueDate"}},
                "drivingLicense.expiryDate": {relation: {entityName: "drivingLicense", refField: "expiryDate"}},
                "drivingLicense.frontImageUrl": {relation: {entityName: "drivingLicense", refField: "frontImageUrl"}},
                "drivingLicense.backImageUrl": {relation: {entityName: "drivingLicense", refField: "backImageUrl"}},
                "drivingLicense.ownerinfoOwnerNic": {relation: {entityName: "drivingLicense", refField: "ownerinfoOwnerNic"}},
                "drivingLicense.issuerinfoId": {relation: {entityName: "drivingLicense", refField: "issuerinfoId"}}
            },
            keyFields: ["ownerNic"],
            joinMetadata: {drivingLicense: {entity: DrivingLicense, fieldName: "drivingLicense", refTable: "DrivingLicense", refColumns: ["ownerinfoOwnerNic"], joinColumns: ["ownerNic"], 'type: psql:ONE_TO_ONE}}
        },
        [ISSUER_INFO]: {
            entityName: "IssuerInfo",
            tableName: "IssuerInfo",
            fieldMetadata: {
                id: {columnName: "id"},
                name: {columnName: "name"},
                issuingAuthority: {columnName: "issuingAuthority"},
                signatureUrl: {columnName: "signatureUrl"},
                "drivingLicenses[].id": {relation: {entityName: "drivingLicenses", refField: "id"}},
                "drivingLicenses[].licenseNumber": {relation: {entityName: "drivingLicenses", refField: "licenseNumber"}},
                "drivingLicenses[].issueDate": {relation: {entityName: "drivingLicenses", refField: "issueDate"}},
                "drivingLicenses[].expiryDate": {relation: {entityName: "drivingLicenses", refField: "expiryDate"}},
                "drivingLicenses[].frontImageUrl": {relation: {entityName: "drivingLicenses", refField: "frontImageUrl"}},
                "drivingLicenses[].backImageUrl": {relation: {entityName: "drivingLicenses", refField: "backImageUrl"}},
                "drivingLicenses[].ownerinfoOwnerNic": {relation: {entityName: "drivingLicenses", refField: "ownerinfoOwnerNic"}},
                "drivingLicenses[].issuerinfoId": {relation: {entityName: "drivingLicenses", refField: "issuerinfoId"}}
            },
            keyFields: ["id"],
            joinMetadata: {drivingLicenses: {entity: DrivingLicense, fieldName: "drivingLicenses", refTable: "DrivingLicense", refColumns: ["issuerinfoId"], joinColumns: ["id"], 'type: psql:MANY_TO_ONE}}
        },
        [VEHICLE_PERMISSION]: {
            entityName: "VehiclePermission",
            tableName: "VehiclePermission",
            fieldMetadata: {
                id: {columnName: "id"},
                vehicleType: {columnName: "vehicleType"},
                issueDate: {columnName: "issueDate"},
                expiryDate: {columnName: "expiryDate"},
                drivinglicenseId: {columnName: "drivinglicenseId"},
                "drivingLicense.id": {relation: {entityName: "drivingLicense", refField: "id"}},
                "drivingLicense.licenseNumber": {relation: {entityName: "drivingLicense", refField: "licenseNumber"}},
                "drivingLicense.issueDate": {relation: {entityName: "drivingLicense", refField: "issueDate"}},
                "drivingLicense.expiryDate": {relation: {entityName: "drivingLicense", refField: "expiryDate"}},
                "drivingLicense.frontImageUrl": {relation: {entityName: "drivingLicense", refField: "frontImageUrl"}},
                "drivingLicense.backImageUrl": {relation: {entityName: "drivingLicense", refField: "backImageUrl"}},
                "drivingLicense.ownerinfoOwnerNic": {relation: {entityName: "drivingLicense", refField: "ownerinfoOwnerNic"}},
                "drivingLicense.issuerinfoId": {relation: {entityName: "drivingLicense", refField: "issuerinfoId"}}
            },
            keyFields: ["id"],
            joinMetadata: {drivingLicense: {entity: DrivingLicense, fieldName: "drivingLicense", refTable: "DrivingLicense", refColumns: ["id"], joinColumns: ["drivinglicenseId"], 'type: psql:ONE_TO_MANY}}
        }
    };

    public isolated function init() returns persist:Error? {
        postgresql:Client|error dbClient = new (host = host, username = user, password = password, database = database, port = port, options = connectionOptions);
        if dbClient is error {
            return <persist:Error>error(dbClient.message());
        }
        self.dbClient = dbClient;
        if defaultSchema != () {
            lock {
                foreach string key in self.metadata.keys() {
                    psql:SQLMetadata metadata = self.metadata.get(key);
                    if metadata.schemaName == () {
                        metadata.schemaName = defaultSchema;
                    }
                    map<psql:JoinMetadata>? joinMetadataMap = metadata.joinMetadata;
                    if joinMetadataMap == () {
                        continue;
                    }
                    foreach [string, psql:JoinMetadata] [_, joinMetadata] in joinMetadataMap.entries() {
                        if joinMetadata.refSchema == () {
                            joinMetadata.refSchema = defaultSchema;
                        }
                    }
                }
            }
        }
        self.persistClients = {
            [DRIVING_LICENSE]: check new (dbClient, self.metadata.get(DRIVING_LICENSE).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS),
            [VEHICLE_CLASS]: check new (dbClient, self.metadata.get(VEHICLE_CLASS).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS),
            [VEHICLE_INFO]: check new (dbClient, self.metadata.get(VEHICLE_INFO).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS),
            [OWNER_INFO]: check new (dbClient, self.metadata.get(OWNER_INFO).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS),
            [ISSUER_INFO]: check new (dbClient, self.metadata.get(ISSUER_INFO).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS),
            [VEHICLE_PERMISSION]: check new (dbClient, self.metadata.get(VEHICLE_PERMISSION).cloneReadOnly(), psql:POSTGRESQL_SPECIFICS)
        };
    }

    isolated resource function get drivinglicenses(DrivingLicenseTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get drivinglicenses/[string id](DrivingLicenseTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post drivinglicenses(DrivingLicenseInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(DRIVING_LICENSE);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from DrivingLicenseInsert inserted in data
            select inserted.id;
    }

    isolated resource function put drivinglicenses/[string id](DrivingLicenseUpdate value) returns DrivingLicense|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(DRIVING_LICENSE);
        }
        _ = check sqlClient.runUpdateQuery(id, value);
        return self->/drivinglicenses/[id].get();
    }

    isolated resource function delete drivinglicenses/[string id]() returns DrivingLicense|persist:Error {
        DrivingLicense result = check self->/drivinglicenses/[id].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(DRIVING_LICENSE);
        }
        _ = check sqlClient.runDeleteQuery(id);
        return result;
    }

    isolated resource function get vehicleclasses(VehicleClassTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get vehicleclasses/[string id](VehicleClassTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post vehicleclasses(VehicleClassInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_CLASS);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from VehicleClassInsert inserted in data
            select inserted.id;
    }

    isolated resource function put vehicleclasses/[string id](VehicleClassUpdate value) returns VehicleClass|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_CLASS);
        }
        _ = check sqlClient.runUpdateQuery(id, value);
        return self->/vehicleclasses/[id].get();
    }

    isolated resource function delete vehicleclasses/[string id]() returns VehicleClass|persist:Error {
        VehicleClass result = check self->/vehicleclasses/[id].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_CLASS);
        }
        _ = check sqlClient.runDeleteQuery(id);
        return result;
    }

    isolated resource function get vehicleinfos(VehicleInfoTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get vehicleinfos/[string id](VehicleInfoTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post vehicleinfos(VehicleInfoInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_INFO);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from VehicleInfoInsert inserted in data
            select inserted.id;
    }

    isolated resource function put vehicleinfos/[string id](VehicleInfoUpdate value) returns VehicleInfo|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_INFO);
        }
        _ = check sqlClient.runUpdateQuery(id, value);
        return self->/vehicleinfos/[id].get();
    }

    isolated resource function delete vehicleinfos/[string id]() returns VehicleInfo|persist:Error {
        VehicleInfo result = check self->/vehicleinfos/[id].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_INFO);
        }
        _ = check sqlClient.runDeleteQuery(id);
        return result;
    }

    isolated resource function get ownerinfos(OwnerInfoTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get ownerinfos/[string ownerNic](OwnerInfoTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post ownerinfos(OwnerInfoInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(OWNER_INFO);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from OwnerInfoInsert inserted in data
            select inserted.ownerNic;
    }

    isolated resource function put ownerinfos/[string ownerNic](OwnerInfoUpdate value) returns OwnerInfo|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(OWNER_INFO);
        }
        _ = check sqlClient.runUpdateQuery(ownerNic, value);
        return self->/ownerinfos/[ownerNic].get();
    }

    isolated resource function delete ownerinfos/[string ownerNic]() returns OwnerInfo|persist:Error {
        OwnerInfo result = check self->/ownerinfos/[ownerNic].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(OWNER_INFO);
        }
        _ = check sqlClient.runDeleteQuery(ownerNic);
        return result;
    }

    isolated resource function get issuerinfos(IssuerInfoTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get issuerinfos/[string id](IssuerInfoTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post issuerinfos(IssuerInfoInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(ISSUER_INFO);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from IssuerInfoInsert inserted in data
            select inserted.id;
    }

    isolated resource function put issuerinfos/[string id](IssuerInfoUpdate value) returns IssuerInfo|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(ISSUER_INFO);
        }
        _ = check sqlClient.runUpdateQuery(id, value);
        return self->/issuerinfos/[id].get();
    }

    isolated resource function delete issuerinfos/[string id]() returns IssuerInfo|persist:Error {
        IssuerInfo result = check self->/issuerinfos/[id].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(ISSUER_INFO);
        }
        _ = check sqlClient.runDeleteQuery(id);
        return result;
    }

    isolated resource function get vehiclepermissions(VehiclePermissionTargetType targetType = <>, sql:ParameterizedQuery whereClause = ``, sql:ParameterizedQuery orderByClause = ``, sql:ParameterizedQuery limitClause = ``, sql:ParameterizedQuery groupByClause = ``) returns stream<targetType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "query"
    } external;

    isolated resource function get vehiclepermissions/[string id](VehiclePermissionTargetType targetType = <>) returns targetType|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor",
        name: "queryOne"
    } external;

    isolated resource function post vehiclepermissions(VehiclePermissionInsert[] data) returns string[]|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_PERMISSION);
        }
        _ = check sqlClient.runBatchInsertQuery(data);
        return from VehiclePermissionInsert inserted in data
            select inserted.id;
    }

    isolated resource function put vehiclepermissions/[string id](VehiclePermissionUpdate value) returns VehiclePermission|persist:Error {
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_PERMISSION);
        }
        _ = check sqlClient.runUpdateQuery(id, value);
        return self->/vehiclepermissions/[id].get();
    }

    isolated resource function delete vehiclepermissions/[string id]() returns VehiclePermission|persist:Error {
        VehiclePermission result = check self->/vehiclepermissions/[id].get();
        psql:SQLClient sqlClient;
        lock {
            sqlClient = self.persistClients.get(VEHICLE_PERMISSION);
        }
        _ = check sqlClient.runDeleteQuery(id);
        return result;
    }

    remote isolated function queryNativeSQL(sql:ParameterizedQuery sqlQuery, typedesc<record {}> rowType = <>) returns stream<rowType, persist:Error?> = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor"
    } external;

    remote isolated function executeNativeSQL(sql:ParameterizedQuery sqlQuery) returns psql:ExecutionResult|persist:Error = @java:Method {
        'class: "io.ballerina.stdlib.persist.sql.datastore.PostgreSQLProcessor"
    } external;

    public isolated function close() returns persist:Error? {
        error? result = self.dbClient.close();
        if result is error {
            return <persist:Error>error(result.message());
        }
        return result;
    }
}

