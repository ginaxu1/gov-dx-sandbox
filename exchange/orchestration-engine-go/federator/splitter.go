package federator

import (
	"regexp"
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/printer"
	"github.com/graphql-go/graphql/language/source"
)

func printCompact(doc *ast.Document) string {
	out := printer.Print(doc).(string)
	// Remove all newlines and compress spaces
	out = strings.ReplaceAll(out, "\n", " ")
	re := regexp.MustCompile(`\s+`)
	out = re.ReplaceAllString(out, " ")
	return strings.TrimSpace(out)
}

func splitQuery(rawQuery string) []*federationServiceRequest {

	// Parse query
	src := source.NewSource(&source.Source{
		Body: []byte(rawQuery),
		Name: "GraphQL request",
	})

	doc, err := parser.Parse(parser.ParseParams{Source: src})
	if err != nil {
		panic(err)
	}

	var results []*federationServiceRequest

	// Traverse top-level definitions
	for _, def := range doc.Definitions {
		if opDef, ok := def.(*ast.OperationDefinition); ok {
			for _, sel := range opDef.SelectionSet.Selections {
				if field, ok := sel.(*ast.Field); ok {
					// Extracting only the provider level queries
					// Check whether the field name matches any registered service
					departmentLevelQuery := &ast.OperationDefinition{
						Operation: ast.OperationTypeQuery,
						Kind:      kinds.OperationDefinition,
						Name:      opDef.Name,
						SelectionSet: &ast.SelectionSet{
							Selections: field.SelectionSet.Selections,
							Kind:       kinds.SelectionSet,
						},
					}

					// Converting the query to a full-featured GraphQL document
					miniDoc := &ast.Document{
						Kind:        kinds.Document,
						Loc:         field.Loc,
						Definitions: []ast.Node{departmentLevelQuery},
					}

					// Creating a federation service request for each department
					results = append(results, &federationServiceRequest{
						ServiceKey: field.Name.Value,
						GraphQLRequest: graphql.Request{
							Query: printCompact(miniDoc),
						},
					})
				}
			}
		}
	}

	return results
}
