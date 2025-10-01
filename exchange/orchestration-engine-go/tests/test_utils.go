package tests

import (
	"testing"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/federator"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/require"
)

// ParseTestQuery is a shared utility function for parsing GraphQL queries in tests
func ParseTestQuery(t *testing.T, query string) *ast.Document {
	src := source.NewSource(&source.Source{
		Body: []byte(query),
		Name: "TestQuery",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err, "Should parse query successfully")
	return doc
}

// CreateTestSchema is a shared utility function for creating test schemas
func CreateTestSchema(t *testing.T) *ast.Document {
	schemaSDL := `
		directive @sourceInfo(
			providerKey: String!
			providerField: String!
		) on FIELD_DEFINITION

		type Query {
			personInfo(nic: String!): PersonInfo
			vehicle(regNo: String!): VehicleInfo
		}

		type PersonInfo {
			fullName: String @sourceInfo(providerKey: "drp", providerField: "person.fullName")
			name: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.name")
			address: String @sourceInfo(providerKey: "drp", providerField: "person.permanentAddress")
			ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data")
			birthInfo: BirthInfo
		}

		type VehicleInfo {
			regNo: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.registrationNumber")
			make: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.make")
			model: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.model")
			maintenanceRecords: [MaintenanceRecord] @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.maintenanceRecords")
		}

		type BirthInfo {
			birthRegistrationNumber: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.brNo")
			birthPlace: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.birthPlace")
			district: String @sourceInfo(providerKey: "rgd", providerField: "getPersonInfo.district")
		}

		type MaintenanceRecord {
			date: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.maintenanceRecords.date")
			description: String @sourceInfo(providerKey: "dmt", providerField: "vehicle.getVehicleInfos.data.maintenanceRecords.description")
		}
	`

	src := source.NewSource(&source.Source{
		Body: []byte(schemaSDL),
		Name: "TestSchema",
	})

	schema, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err, "Should parse schema successfully")
	return schema
}

// PushValue is a wrapper around federator.PushValue for tests
func PushValue(obj interface{}, path string, value interface{}) (interface{}, error) {
	return federator.PushValue(obj, path, value)
}

// GetValueAtPath is a wrapper around federator.GetValueAtPath for tests
func GetValueAtPath(data interface{}, path string) (interface{}, error) {
	return federator.GetValueAtPath(data, path)
}
