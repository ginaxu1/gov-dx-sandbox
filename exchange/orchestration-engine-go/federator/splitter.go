package federator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/printer"
	"github.com/graphql-go/graphql/language/source"
)

// this file is responsible for splitting the incoming request into multiple requests based on the serviceKeys

// Arriving GraphQL Request:
// {
//   "query": "query MyQuery { drp { person(nic: \"199512345678\") { nic photo } } dmt { vehicle { getVehicleInfos { data { model } } } } }",
//   "variables": null
// }

// Split into multiple requests:
// [
//   {
//     "serviceKey": "drp",
//     "graphqlQuery": {
//       "query": "query MyQuery { person(nic: \"199512345678\") { nic photo } }",
//       "variables": null
//     }
//   },
//   {
//     "serviceKey": "dmt",
//     "graphqlQuery": {
//       "query": "query MyQuery { vehicle { getVehicleInfos { data { model } } } }",
//       "variables": null
//     }
//   }
// ]

func printCompact(doc *ast.Document) string {
	out := printer.Print(doc).(string)
	// Remove all newlines and compress spaces
	out = strings.ReplaceAll(out, "\n", " ")
	re := regexp.MustCompile(`\s+`)
	out = re.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}

func splitQuery(rawQuery string) []*FederationServiceRequest {

	// Parse query
	src := source.NewSource(&source.Source{
		Body: []byte(rawQuery),
		Name: "GraphQL request",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		panic(err)
	}

	var results []*FederationServiceRequest

	// Traverse top-level definitions
	for _, def := range doc.Definitions {
		if opDef, ok := def.(*ast.OperationDefinition); ok {
			for _, sel := range opDef.SelectionSet.Selections {
				if field, ok := sel.(*ast.Field); ok {
					// Build a mini query with only this field
					newOp := &ast.OperationDefinition{
						Operation: ast.OperationTypeQuery,
						Kind:      "OperationDefinition",
						Name:      opDef.Name,
						SelectionSet: &ast.SelectionSet{
							Selections: field.SelectionSet.Selections,
							Kind:       "SelectionSet",
						},
					}

					miniDoc := &ast.Document{
						Kind:        "Document",
						Loc:         field.Loc,
						Definitions: []ast.Node{newOp},
					}

					fmt.Println("----- Subquery -----")
					fmt.Println(printCompact(doc))
					results = append(results, &FederationServiceRequest{
						ServiceKey: field.Name.Value,
						GraphqlQuery: GraphQLRequest{
							Query: printCompact(miniDoc),
						},
					})
				}
			}
		}
	}

	return results
}
