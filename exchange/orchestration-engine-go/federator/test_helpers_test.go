package federator

import (
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
	"github.com/stretchr/testify/require"
)

const defaultTestSchemaSDL = `
directive @sourceInfo(
	providerKey: String!
	schemaId: String
	providerField: String!
) on FIELD_DEFINITION

type Query {
	personInfo(nic: String!): PersonInfo
	personInfos(nics: [String!]!): [PersonInfo]
}

type PersonInfo {
	fullName: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.fullName")
	name: String @sourceInfo(providerKey: "rgd", schemaId: "rgd-schema-v1", providerField: "getPersonInfo.name")
	address: String @sourceInfo(providerKey: "drp", schemaId: "drp-schema-v1", providerField: "person.permanentAddress")
	ownedVehicles: [VehicleInfo] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data")
	class: [VehicleClass] @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.classes")
}

type VehicleInfo {
	regNo: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.registrationNumber")
	make: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.make")
	model: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.data.model")
}

type VehicleClass {
	className: String @sourceInfo(providerKey: "dmt", schemaId: "dmt-schema-v1", providerField: "vehicle.getVehicleInfos.classes.className")
}
`

func mustParseDocument(t *testing.T, input string) *ast.Document {
	t.Helper()

	src := source.NewSource(&source.Source{
		Body: []byte(input),
		Name: "TestDoc",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	require.NoError(t, err)
	return doc
}

func ParseTestQuery(t *testing.T, query string) *ast.Document {
	return mustParseDocument(t, query)
}

func ParseSchemaDoc(t *testing.T, sdl string) *ast.Document {
	return mustParseDocument(t, sdl)
}

func ParseQueryDoc(t *testing.T, query string) *ast.Document {
	return mustParseDocument(t, query)
}

func CreateTestSchema(t *testing.T) *ast.Document {
	return ParseSchemaDoc(t, defaultTestSchemaSDL)
}
