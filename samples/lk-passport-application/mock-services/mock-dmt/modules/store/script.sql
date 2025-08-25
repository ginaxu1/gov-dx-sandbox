-- AUTO-GENERATED FILE.

-- This file is an auto-generated file by Ballerina persistence layer for model.
-- Please verify the generated scripts and execute them against the target DB server.

DROP TABLE IF EXISTS "VehicleInfo";
DROP TABLE IF EXISTS "VehiclePermission";
DROP TABLE IF EXISTS "DrivingLicense";
DROP TABLE IF EXISTS "IssuerInfo";
DROP TABLE IF EXISTS "OwnerInfo";
DROP TABLE IF EXISTS "VehicleClass";

CREATE TABLE "VehicleClass" (
	"id" VARCHAR(191) NOT NULL,
	"className" VARCHAR(191) NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE "OwnerInfo" (
	"ownerNic" VARCHAR(191) NOT NULL,
	"name" VARCHAR(191) NOT NULL,
	"address" VARCHAR(191) NOT NULL,
	"birthDate" VARCHAR(191) NOT NULL,
	"signatureUrl" VARCHAR(191) NOT NULL,
	"bloodGroup" VARCHAR(11) CHECK ("bloodGroup" IN ('A_POSITIVE', 'A_NEGATIVE', 'B_POSITIVE', 'B_NEGATIVE', 'AB_POSITIVE', 'AB_NEGATIVE', 'O_POSITIVE', 'O_NEGATIVE')) NOT NULL,
	PRIMARY KEY("ownerNic")
);

CREATE TABLE "IssuerInfo" (
	"id" VARCHAR(191) NOT NULL,
	"name" VARCHAR(191) NOT NULL,
	"issuingAuthority" VARCHAR(191) NOT NULL,
	"signatureUrl" VARCHAR(191) NOT NULL,
	PRIMARY KEY("id")
);

CREATE TABLE "DrivingLicense" (
	"id" VARCHAR(191) NOT NULL,
	"licenseNumber" VARCHAR(191) NOT NULL,
	"issueDate" VARCHAR(191) NOT NULL,
	"expiryDate" VARCHAR(191) NOT NULL,
	"frontImageUrl" VARCHAR(191) NOT NULL,
	"backImageUrl" VARCHAR(191) NOT NULL,
	"ownerinfoOwnerNic" VARCHAR(191) UNIQUE NOT NULL,
	FOREIGN KEY("ownerinfoOwnerNic") REFERENCES "OwnerInfo"("ownerNic"),
	"issuerinfoId" VARCHAR(191) NOT NULL,
	FOREIGN KEY("issuerinfoId") REFERENCES "IssuerInfo"("id"),
	PRIMARY KEY("id")
);

CREATE TABLE "VehiclePermission" (
	"id" VARCHAR(191) NOT NULL,
	"vehicleType" VARCHAR(2) CHECK ("vehicleType" IN ('A1', 'A', 'B', 'C1', 'C', 'CE', 'D1', 'D', 'DE', 'G1', 'G', 'J')) NOT NULL,
	"issueDate" VARCHAR(191) NOT NULL,
	"expiryDate" VARCHAR(191) NOT NULL,
	"drivinglicenseId" VARCHAR(191) NOT NULL,
	FOREIGN KEY("drivinglicenseId") REFERENCES "DrivingLicense"("id"),
	PRIMARY KEY("id")
);

CREATE TABLE "VehicleInfo" (
	"id" VARCHAR(191) NOT NULL,
	"make" VARCHAR(191) NOT NULL,
	"model" VARCHAR(191) NOT NULL,
	"yearOfManufacture" INT NOT NULL,
	"ownerNic" VARCHAR(191) NOT NULL,
	"engineNumber" VARCHAR(191) NOT NULL,
	"conditionAndNotes" VARCHAR(191) NOT NULL,
	"registrationNumber" VARCHAR(191) NOT NULL,
	"vehicleclassId" VARCHAR(191) NOT NULL,
	FOREIGN KEY("vehicleclassId") REFERENCES "VehicleClass"("id"),
	PRIMARY KEY("id")
);


