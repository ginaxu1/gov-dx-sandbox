package federator

import (
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func BuildProviderLevelQuery(fieldsMap []string) []*FederationServiceAST {
	var queries []*FederationServiceAST
	var addedServiceKeys []string

	for _, field := range fieldsMap {
		// split with periods and create a string.
		var args = strings.Split(field, ".")

		serviceKey := args[0]

		if !contains(addedServiceKeys, serviceKey) {
			addedServiceKeys = append(addedServiceKeys, serviceKey)
			queries = append(queries, &FederationServiceAST{
				ServiceKey: serviceKey,
				QueryAst: &ast.Document{
					Kind: kinds.Document,
					Definitions: []ast.Node{
						&ast.OperationDefinition{
							Kind:      kinds.OperationDefinition,
							Operation: "query",
							Name: &ast.Name{
								Kind:  kinds.Name,
								Value: "Query" + serviceKey,
							},
							SelectionSet: &ast.SelectionSet{
								Kind: kinds.SelectionSet,
							},
						},
					},
				},
			})
		}
		// find the query with the service key
		for _, q := range queries {
			if q.ServiceKey == serviceKey {
				pushFieldToAst(args[1:], q.QueryAst.Definitions[0].(*ast.OperationDefinition).SelectionSet)
				break
			}
		}
	}
	return queries
}

// BuildArrayProviderQuery creates a provider query specifically for array response fields
func BuildArrayProviderQuery(fieldsMap []string, arrayFields []string) []*FederationServiceAST {
	var queries []*FederationServiceAST
	var addedServiceKeys []string

	// Process regular fields first
	for _, field := range fieldsMap {
		var args = strings.Split(field, ".")
		serviceKey := args[0]

		if !contains(addedServiceKeys, serviceKey) {
			addedServiceKeys = append(addedServiceKeys, serviceKey)
			queries = append(queries, &FederationServiceAST{
				ServiceKey: serviceKey,
				QueryAst: &ast.Document{
					Kind: kinds.Document,
					Definitions: []ast.Node{
						&ast.OperationDefinition{
							Kind:      kinds.OperationDefinition,
							Operation: "query",
							Name: &ast.Name{
								Kind:  kinds.Name,
								Value: "Query" + serviceKey,
							},
							SelectionSet: &ast.SelectionSet{
								Kind: kinds.SelectionSet,
							},
						},
					},
				},
			})
		}

		// Add field to query
		for _, q := range queries {
			if q.ServiceKey == serviceKey {
				pushFieldToAst(args[1:], q.QueryAst.Definitions[0].(*ast.OperationDefinition).SelectionSet)
				break
			}
		}
	}

	// Process array fields
	for _, field := range arrayFields {
		var args = strings.Split(field, ".")
		serviceKey := args[0]

		if !contains(addedServiceKeys, serviceKey) {
			addedServiceKeys = append(addedServiceKeys, serviceKey)
			queries = append(queries, &FederationServiceAST{
				ServiceKey: serviceKey,
				QueryAst: &ast.Document{
					Kind: kinds.Document,
					Definitions: []ast.Node{
						&ast.OperationDefinition{
							Kind:      kinds.OperationDefinition,
							Operation: "query",
							Name: &ast.Name{
								Kind:  kinds.Name,
								Value: "Query" + serviceKey,
							},
							SelectionSet: &ast.SelectionSet{
								Kind: kinds.SelectionSet,
							},
						},
					},
				},
			})
		}

		// Add array field to query
		for _, q := range queries {
			if q.ServiceKey == serviceKey {
				pushArrayFieldToAst(args[1:], q.QueryAst.Definitions[0].(*ast.OperationDefinition).SelectionSet)
				break
			}
		}
	}

	return queries
}

func pushFieldToAst(field []string, parentField *ast.SelectionSet) {
	// loop through selectionSets to match the field path.
	var current = field[0]

	// Cases:
	// 1. If there are no fields, add the field to the selection set.
	// 2. If there are fields, check if the field already exists.
	//    a. If it exists, go deeper if there are more fields nested.
	//    b. If it does not exist, add it to the selection set and go deeper if there are more fields nested.

	// if there are no fields, add the field to the selection set.
	if len(parentField.Selections) == 0 {
		parentField.Selections = append(parentField.Selections, &ast.Field{
			Kind: kinds.Field,
			Name: &ast.Name{
				Kind:  kinds.Name,
				Value: current,
			},
		})
		// if there are more fields nested, go deeper.
		if len(field) > 1 {
			parentField.Selections[len(parentField.Selections)-1].(*ast.Field).SelectionSet = &ast.SelectionSet{
				Kind: kinds.SelectionSet,
			}
			pushFieldToAst(field[1:], parentField.Selections[0].(*ast.Field).GetSelectionSet())
		}
		return
	} else {
		// if there are fields, check if the field already exists.
		var found bool
		for _, f := range parentField.Selections {
			if current == f.(*ast.Field).Name.Value {
				found = true
				// if there are more fields nested, go deeper.
				if len(field) > 1 {
					if f.GetSelectionSet() == nil {
						f.(*ast.Field).SelectionSet = &ast.SelectionSet{
							Kind: kinds.SelectionSet,
						}
					}
					pushFieldToAst(field[1:], f.GetSelectionSet())
				}
				break
			}
		}
		if !found {
			var newField = &ast.Field{
				Kind: kinds.Field,
				Name: &ast.Name{
					Kind:  kinds.Name,
					Value: current,
				},
			}
			// if the field does not exist, add it to the selection set.
			parentField.Selections = append(parentField.Selections, newField)

			// if there are more fields nested, go deeper.
			if len(field) > 1 {
				newField.SelectionSet = &ast.SelectionSet{
					Kind: kinds.SelectionSet,
				}
				pushFieldToAst(field[1:], parentField.Selections[len(parentField.Selections)-1].(*ast.Field).GetSelectionSet())
			}
			return
		}
	}
}

// pushArrayFieldToAst handles array fields specifically, ensuring proper array structure
func pushArrayFieldToAst(field []string, parentField *ast.SelectionSet) {
	// For array fields, we need to handle the structure differently
	var current = field[0]

	// Check if field already exists
	var found bool
	for _, f := range parentField.Selections {
		if current == f.(*ast.Field).Name.Value {
			found = true
			// If there are more fields nested, go deeper
			if len(field) > 1 {
				if f.GetSelectionSet() == nil {
					f.(*ast.Field).SelectionSet = &ast.SelectionSet{
						Kind: kinds.SelectionSet,
					}
				}
				pushArrayFieldToAst(field[1:], f.GetSelectionSet())
			}
			break
		}
	}

	if !found {
		var newField = &ast.Field{
			Kind: kinds.Field,
			Name: &ast.Name{
				Kind:  kinds.Name,
				Value: current,
			},
		}

		// Add the field to the selection set
		parentField.Selections = append(parentField.Selections, newField)

		// If there are more fields nested, go deeper
		if len(field) > 1 {
			newField.SelectionSet = &ast.SelectionSet{
				Kind: kinds.SelectionSet,
			}
			pushArrayFieldToAst(field[1:], newField.SelectionSet)
		}
	}
}

// isArrayField checks if a field path represents an array field based on the schema
func isArrayFieldInSchema(fieldPath string, schema *ast.Document) bool {
	// This function would need to be implemented to check the schema
	// for array type definitions. For now, we'll use a simple heuristic.
	return strings.Contains(fieldPath, "[]") || strings.Contains(fieldPath, "array")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
