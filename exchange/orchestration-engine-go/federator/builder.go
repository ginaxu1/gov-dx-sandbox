package federator

import (
	"sort"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
)

func flattenedPaths(field *ast.Field, parentPath string) []string {
	// Initialize a slice to hold the flattened paths for this field's sub-tree.
	var paths []string

	// If the field has no selection set, it's a leaf node.
	// We construct its full path and return it in a single-element slice.
	if field.SelectionSet == nil {
		if parentPath == "" {
			return []string{field.Name.Value}
		}
		// Construct the full path by appending the field name to the parent path.
		fullPath := parentPath + "." + field.Name.Value
		return []string{fullPath}
	}

	// This field is an internal node, so we need to process its sub-selections.
	// First, determine the current path prefix for this level of recursion.
	currentPath := field.Name.Value
	if parentPath != "" {
		currentPath = parentPath + "." + field.Name.Value
	}

	// Iterate over each selection within the field's selection set.
	for _, selection := range field.SelectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			// Recursively call flattenedPaths on the sub-field.
			// The currentPath serves as the parent path for the next level.
			subPaths := flattenedPaths(sel, currentPath)

			// Append the returned sub-paths to our main paths slice.
			// The `...` operator unpacks the sub-paths slice,
			// appending each element individually.
			paths = append(paths, subPaths...)
		default:
			continue
		}
	}

	return paths
}

// the function iterates through a list of keys of a lookup table
// and returns a list of keys that match the given field paths.
func matchFieldPaths(fieldPaths []string, lookupTable map[string]string) []string {
	var matchedKeys []string
	for key := range lookupTable {
		for _, path := range fieldPaths {
			if key == path {
				matchedKeys = append(matchedKeys, lookupTable[key])
				break // No need to check other paths once a match is found
			}
		}
	}

	sort.Strings(matchedKeys)

	return matchedKeys
}

func BuildProviderLevelQuery(fieldsMap []string) []*federationServiceAST {
	var queries []*federationServiceAST
	var addedServiceKeys []string

	for _, field := range fieldsMap {
		// split with periods and create a string.
		var args = strings.Split(field, ".")

		serviceKey := args[0]

		if !contains(addedServiceKeys, serviceKey) {
			addedServiceKeys = append(addedServiceKeys, serviceKey)
			queries = append(queries, &federationServiceAST{
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
