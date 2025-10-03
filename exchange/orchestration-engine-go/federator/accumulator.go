package federator

import (
	"fmt"
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/visitor"
)

func AccumulateResponse(queryAST *ast.Document, federatedResponse *FederationResponse) graphql.Response {
	// Use the simple accumulator for backward compatibility
	// For array-aware processing, use AccumulateResponseWithSchemaInfo instead
	return accumulateResponseSimple(queryAST, federatedResponse)
}

// accumulateResponseSimple is the fallback simple accumulator
func accumulateResponseSimple(queryAST *ast.Document, federatedResponse *FederationResponse) graphql.Response {
	responseData := make(map[string]interface{})
	var path = make([]string, 0)
	var isTopLevel = true

	visitor.Visit(queryAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value

				// Handle top-level query fields
				if isTopLevel {
					responseData[fieldName] = make(map[string]interface{})
					path = append(path, fieldName)
					isTopLevel = false
					return visitor.ActionNoChange, p.Node
				}

				// Handle nested fields
				if len(path) > 0 {
					// Check if this is a nested field of an array (skip processing)
					if isNestedFieldOfArray(path, queryAST) {
						return visitor.ActionNoChange, p.Node
					}

					var providerInfo = federator.ExtractSourceInfoFromDirective(node)
					if providerInfo != nil {
						var response = federatedResponse.GetProviderResponse(providerInfo.ProviderKey)
						if response != nil {
							var value, err = GetValueAtPath(response.Response.Data, providerInfo.ProviderField)
							if err == nil {
								// Check if this is an array field by looking at the selection set
								if node.SelectionSet != nil && len(node.SelectionSet.Selections) > 0 {
									// This is an array field with nested selections
									processArrayFieldSimple(responseData, path, fieldName, value, node.SelectionSet, federatedResponse)
								} else {
									// Simple field
									fullPath := strings.Join(append(path, fieldName), ".")
									_, err = PushValue(responseData, fullPath, value)
									if err != nil {
										fmt.Printf("Error pushing value at path %s: %v\n", fullPath, err)
									}
								}
							} else {
								fmt.Printf("Error getting value at path %s: %v\n", providerInfo.ProviderField, err)
							}
						}
					} else {
						fmt.Printf("Warning: No @sourceInfo directive found for field '%s' at path '%s'. Skipping field.\n", fieldName, strings.Join(append(path, fieldName), "."))
					}
				}
				path = append(path, fieldName)
			}
			return visitor.ActionNoChange, p.Node
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value
				if len(path) > 0 && path[len(path)-1] == fieldName {
					path = path[:len(path)-1]
				}
				if len(path) == 0 {
					isTopLevel = true
				}
			}
			return visitor.ActionNoChange, p.Node
		},
	}, nil)

	return graphql.Response{Data: responseData}
}

// accumulateResponseWithSchema uses schema info to handle arrays properly
func accumulateResponseWithSchema(queryAST *ast.Document, federatedResponse *FederationResponse, schemaInfoMap map[string]*SourceSchemaInfo) graphql.Response {
	responseData := make(map[string]interface{})
	var path = make([]string, 0)
	var isTopLevel = true

	visitor.Visit(queryAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value

				// Handle top-level query fields
				if isTopLevel {
					responseData[fieldName] = make(map[string]interface{})
					path = append(path, fieldName)
					isTopLevel = false
					return visitor.ActionNoChange, p.Node
				}

				// Handle nested fields
				if len(path) > 0 {
					fullPath := strings.Join(append(path, fieldName), ".")
					schemaInfo, exists := schemaInfoMap[fullPath]

					if exists {
						if schemaInfo.IsArray {
							// Handle array field
							processArrayFieldWithSchema(responseData, path, fieldName, schemaInfo, federatedResponse)
						} else {
							// Handle simple field
							processSimpleField(responseData, path, fieldName, schemaInfo, federatedResponse)
						}
					} else {
						fmt.Printf("Warning: No schema info found for field '%s' at path '%s'. Skipping field.\n", fieldName, fullPath)
					}
				}
				path = append(path, fieldName)
			}
			return visitor.ActionNoChange, p.Node
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value
				if len(path) > 0 && path[len(path)-1] == fieldName {
					path = path[:len(path)-1]
				}
				if len(path) == 0 {
					isTopLevel = true
				}
			}
			return visitor.ActionNoChange, p.Node
		},
	}, nil)

	return graphql.Response{Data: responseData}
}

// isNestedFieldOfArray checks if the current field is a nested field of an array
// by examining the query AST to determine which fields are arrays
func isNestedFieldOfArray(path []string, queryAST *ast.Document) bool {
	if len(path) < 2 {
		return false
	}

	// Get the parent field (the array field) from the path
	arrayFieldName := path[len(path)-2]

	// Find the array field in the query AST
	arrayField := findFieldInQuery(queryAST, arrayFieldName)
	if arrayField == nil {
		return false
	}

	// Check if this field has a selection set (indicating it's an array/object field)
	// and if we're currently processing a nested field of it
	return arrayField.SelectionSet != nil && len(arrayField.SelectionSet.Selections) > 0
}

// findFieldInQuery recursively searches for a field with the given name in the query AST
func findFieldInQuery(queryAST *ast.Document, fieldName string) *ast.Field {
	for _, definition := range queryAST.Definitions {
		if operation, ok := definition.(*ast.OperationDefinition); ok {
			if operation.SelectionSet != nil {
				for _, selection := range operation.SelectionSet.Selections {
					if field, ok := selection.(*ast.Field); ok {
						if field.Name != nil && field.Name.Value == fieldName {
							return field
						}
						// Recursively search in nested selections
						if found := findFieldInSelectionSet(field.SelectionSet, fieldName); found != nil {
							return found
						}
					}
				}
			}
		}
	}
	return nil
}

// findFieldInSelectionSet recursively searches for a field in a selection set
func findFieldInSelectionSet(selectionSet *ast.SelectionSet, fieldName string) *ast.Field {
	if selectionSet == nil {
		return nil
	}

	for _, selection := range selectionSet.Selections {
		if field, ok := selection.(*ast.Field); ok {
			if field.Name != nil && field.Name.Value == fieldName {
				return field
			}
			// Recursively search in nested selections
			if found := findFieldInSelectionSet(field.SelectionSet, fieldName); found != nil {
				return found
			}
		}
	}
	return nil
}

// processArrayFieldSimple handles array fields with nested selections
func processArrayFieldSimple(responseData map[string]interface{}, path []string, fieldName string, sourceArray interface{}, selectionSet *ast.SelectionSet, federatedResponse *FederationResponse) {
	// Convert source array to []interface{}
	var arrayData []interface{}
	if arr, ok := sourceArray.([]interface{}); ok {
		arrayData = arr
	} else {
		fmt.Printf("Expected array at field %s, got %T\n", fieldName, sourceArray)
		return
	}

	// Create destination array
	destinationArray := make([]map[string]interface{}, 0, len(arrayData))

	// Process each item in the source array
	for _, sourceItem := range arrayData {
		if sourceItemMap, ok := sourceItem.(map[string]interface{}); ok {
			// Create destination object
			destinationObject := make(map[string]interface{})

			// Process each nested field in the selection set
			for _, selection := range selectionSet.Selections {
				if nestedField, ok := selection.(*ast.Field); ok {
					nestedFieldName := nestedField.Name.Value

					// Get the @sourceInfo directive for the nested field
					var nestedProviderInfo = federator.ExtractSourceInfoFromDirective(nestedField)
					if nestedProviderInfo != nil {
						// Extract the relative field path from the full provider field path
						relativeFieldPath := extractRelativeFieldPath(nestedProviderInfo.ProviderField)

						// Get value from source item using relative field path
						value, err := GetValueAtPath(sourceItemMap, relativeFieldPath)
						if err == nil {
							destinationObject[nestedFieldName] = value
						} else {
							fmt.Printf("Error getting sub-field %s from source item: %v\n", relativeFieldPath, err)
						}
					} else {
						fmt.Printf("Warning: No @sourceInfo directive found for nested field '%s'\n", nestedFieldName)
					}
				}
			}

			destinationArray = append(destinationArray, destinationObject)
		}
	}

	// Add the completed array to the response
	fullPath := strings.Join(append(path, fieldName), ".")
	_, err := PushValue(responseData, fullPath, destinationArray)
	if err != nil {
		fmt.Printf("Error pushing array at path %s: %v\n", fullPath, err)
	}
}

// extractRelativeFieldPath extracts the relative field path from a full provider field path
func extractRelativeFieldPath(providerField string) string {
	// For paths like "vehicle.getVehicleInfos.data.registrationNumber",
	// extract just "registrationNumber"
	parts := strings.Split(providerField, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return providerField
}

// processSimpleField handles simple (non-array) fields
func processSimpleField(responseData map[string]interface{}, path []string, fieldName string, schemaInfo *SourceSchemaInfo, federatedResponse *FederationResponse) {
	response := federatedResponse.GetProviderResponse(schemaInfo.ProviderKey)
	if response != nil {
		value, err := GetValueAtPath(response.Response.Data, schemaInfo.ProviderField)
		if err == nil {
			fullPath := strings.Join(append(path, fieldName), ".")
			_, err = PushValue(responseData, fullPath, value)
			if err != nil {
				fmt.Printf("Error pushing value at path %s: %v\n", fullPath, err)
			}
		} else {
			fmt.Printf("Error getting value at path %s: %v\n", schemaInfo.ProviderField, err)
		}
	}
}

// processArrayFieldWithSchema handles array fields using schema information
func processArrayFieldWithSchema(responseData map[string]interface{}, path []string, fieldName string, schemaInfo *SourceSchemaInfo, federatedResponse *FederationResponse) {
	response := federatedResponse.GetProviderResponse(schemaInfo.ProviderKey)
	if response == nil {
		fmt.Printf("No response found for provider %s\n", schemaInfo.ProviderKey)
		return
	}

	// Get the source array from the provider response
	sourceArray, err := GetValueAtPath(response.Response.Data, schemaInfo.ProviderArrayFieldPath)
	if err != nil {
		fmt.Printf("Error getting array at path %s: %v\n", schemaInfo.ProviderArrayFieldPath, err)
		return
	}

	// Convert to array if it's not already
	var arrayData []interface{}
	if arr, ok := sourceArray.([]interface{}); ok {
		arrayData = arr
	} else {
		fmt.Printf("Expected array at path %s, got %T\n", schemaInfo.ProviderArrayFieldPath, sourceArray)
		return
	}

	// Create destination array
	destinationArray := make([]map[string]interface{}, 0, len(arrayData))

	// Process each item in the source array
	for _, sourceItem := range arrayData {
		if sourceItemMap, ok := sourceItem.(map[string]interface{}); ok {
			// Create destination object
			destinationObject := make(map[string]interface{})

			// Process each sub-field
			for subFieldName, subFieldSchemaInfo := range schemaInfo.SubFieldSchemaInfos {
				// Get value from source item using relative field path
				value, err := GetValueAtPath(sourceItemMap, subFieldSchemaInfo.ProviderField)
				if err == nil {
					destinationObject[subFieldName] = value
				} else {
					fmt.Printf("Error getting sub-field %s from source item: %v\n", subFieldSchemaInfo.ProviderField, err)
				}
			}

			destinationArray = append(destinationArray, destinationObject)
		}
	}

	// Add the completed array to the response
	fullPath := strings.Join(append(path, fieldName), ".")
	_, err = PushValue(responseData, fullPath, destinationArray)
	if err != nil {
		fmt.Printf("Error pushing array at path %s: %v\n", fullPath, err)
	}
}

// AccumulateResponseWithSchemaInfo uses schema information for array-aware processing
func AccumulateResponseWithSchemaInfo(queryAST *ast.Document, federatedResponse *FederationResponse, schemaInfoMap map[string]*SourceSchemaInfo) graphql.Response {
	responseData := make(map[string]interface{})

	// Process each field in the schema info map
	for fieldPath, schemaInfo := range schemaInfoMap {
		if schemaInfo.IsArray {
			// Handle array fields with object-by-object processing
			err := accumulateArrayResponse(responseData, fieldPath, schemaInfo, federatedResponse)
			if err != nil {
				fmt.Printf("Error processing array field %s: %v\n", fieldPath, err)
			}
		} else {
			// Handle regular fields
			response := federatedResponse.GetProviderResponse(schemaInfo.ProviderKey)
			if response != nil {
				value, err := GetValueAtPath(response.Response.Data, schemaInfo.ProviderField)
				if err == nil {
					_, err = PushValue(responseData, fieldPath, value)
					if err != nil {
						fmt.Printf("Error pushing value at path %s: %v\n", fieldPath, err)
					}
				} else {
					fmt.Printf("Error getting value at path %s: %v\n", schemaInfo.ProviderField, err)
				}
			}
		}
	}

	return graphql.Response{
		Data: responseData,
	}
}

// accumulateArrayResponse handles the logic for building an array of objects from a provider response
func accumulateArrayResponse(
	destination map[string]interface{},
	fieldPath string, // e.g., "personInfo.ownedVehicles"
	fieldSchemaInfo *SourceSchemaInfo, // The schema info for the 'ownedVehicles' field
	federatedResponse *FederationResponse,
) error {
	// 1. Get the provider response
	response := federatedResponse.GetProviderResponse(fieldSchemaInfo.ProviderKey)
	if response == nil {
		return fmt.Errorf("no response found for provider %s", fieldSchemaInfo.ProviderKey)
	}

	// 2. Extract the entire source array from the provider response
	// Uses the ProviderArrayFieldPath from the schema info.
	sourceArrayInterface, err := GetValueAtPath(response.Response.Data, fieldSchemaInfo.ProviderArrayFieldPath)
	if err != nil {
		// Handle cases where the path doesn't exist gracefully
		return fmt.Errorf("source array path not found: %s", fieldSchemaInfo.ProviderArrayFieldPath)
	}

	sourceArray, ok := sourceArrayInterface.([]interface{})
	if !ok {
		// The data at the path was not an array, which is an error
		return fmt.Errorf("expected an array at path %s but got %T", fieldSchemaInfo.ProviderArrayFieldPath, sourceArrayInterface)
	}

	// 3. Create the destination array that we will populate
	destinationArray := make([]map[string]interface{}, 0, len(sourceArray))

	// 4. Iterate over each item in the source array
	for _, sourceItemInterface := range sourceArray {
		sourceItem, ok := sourceItemInterface.(map[string]interface{})
		if !ok {
			// Log a warning if an item in the array is not an object
			fmt.Printf("Warning: Expected object in array at %s, got %T\n", fieldSchemaInfo.ProviderArrayFieldPath, sourceItemInterface)
			continue
		}

		// 5. Create a new destination object for each source item
		destinationObject := make(map[string]interface{})

		// 6. Populate the destination object using the sub-field mappings
		for consumerFieldName, subFieldInfo := range fieldSchemaInfo.SubFieldSchemaInfos {
			// The provider field path (e.g., "registrationNumber") is relative to the source item
			value, err := GetValueAtPath(sourceItem, subFieldInfo.ProviderField)
			if err == nil {
				// Use the final part of the consumer field name as the key (e.g., "regNo")
				keyParts := strings.Split(consumerFieldName, ".")
				key := keyParts[len(keyParts)-1]
				destinationObject[key] = value
			} else {
				// Field not found in source item, skip it silently
			}
		}
		destinationArray = append(destinationArray, destinationObject)
	}

	// 7. Push the completed destination array into the final response structure
	_, err = PushValue(destination, fieldPath, destinationArray)
	return err
}

// AccumulateArrayResponse handles array response accumulation specifically
func AccumulateArrayResponse(queryAST *ast.Document, federatedResponse *FederationResponse) graphql.Response {
	// Traverse the query AST and accumulate data from federatedResponse
	// into a single response structure with enhanced array handling.
	responseData := make(map[string]interface{})

	var path = make([]string, 0)
	var isTopLevel = true

	visitor.Visit(queryAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value

				// Handle top-level query fields
				if isTopLevel {
					// Initialize the top-level field structure
					responseData[fieldName] = make(map[string]interface{})
					path = append(path, fieldName)
					isTopLevel = false
					return visitor.ActionNoChange, p.Node
				}

				// Handle nested fields
				if len(path) > 0 {
					var providerInfo = federator.ExtractSourceInfoFromDirective(node)

					if providerInfo != nil {
						var response = federatedResponse.GetProviderResponse(providerInfo.ProviderKey)

						if response != nil {
							var value, err = GetValueAtPath(response.Response.Data, providerInfo.ProviderField)
							if err == nil {
								// Enhanced array handling
								_, err = PushArrayValue(responseData, strings.Join(append(path, fieldName), "."), value)
								if err != nil {
									fmt.Printf("Error pushing array value at path %s: %v\n", strings.Join(path, "."), err)
								}
							} else {
								fmt.Printf("Error getting value at path %s: %v\n", providerInfo.ProviderField, err)
							}
						}
					}
				}

				path = append(path, fieldName)
			}
			return visitor.ActionNoChange, p.Node
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			// remove from path
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value
				if len(path) > 0 && path[len(path)-1] == fieldName {
					path = path[:len(path)-1]
				}
				// Reset top-level flag when leaving the first field
				if len(path) == 0 {
					isTopLevel = true
				}
			}
			return visitor.ActionNoChange, p.Node
		},
	}, nil)

	return graphql.Response{
		Data: responseData,
	}
}

// PushValue pushes a value into a JSON-like structure (map[string]interface{} / []interface{})
// using a dot-notation path. If a segment already points to an array, the value is appended to all items.
func PushValue(obj interface{}, path string, value interface{}) (interface{}, error) {
	keys := strings.Split(path, ".")
	return pushRecursive(obj, keys, value)
}

func pushRecursive(obj interface{}, keys []string, value interface{}) (interface{}, error) {
	if len(keys) == 0 {
		return value, nil
	}

	key := keys[0]

	switch curr := obj.(type) {
	case map[string]interface{}:
		child, ok := curr[key]
		if !ok {
			// If more keys → create map, else assign value
			if len(keys) > 1 {
				child = map[string]interface{}{}
			} else {
				curr[key] = value
				return curr, nil
			}
		}

		newChild, err := pushRecursive(child, keys[1:], value)
		if err != nil {
			return nil, err
		}
		curr[key] = newChild
		return curr, nil

	case []interface{}:
		// For arrays: apply pushRecursive to all elements
		newArr := make([]interface{}, len(curr))
		for i, elem := range curr {
			newChild, err := pushRecursive(elem, keys, value)
			if err != nil {
				return nil, err
			}
			newArr[i] = newChild
		}
		return newArr, nil

	case nil:
		// Initialize a map if nil
		child := map[string]interface{}{}
		newChild, err := pushRecursive(child, keys, value)
		if err != nil {
			return nil, err
		}
		return newChild, nil

	default:
		return nil, fmt.Errorf("unexpected type %T at key %q", obj, key)
	}
}

func GetValueAtPath(data interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	return getValueRecursive(data, keys)
}

// isArrayFieldValue checks if a field is an array field based on the value type
func isArrayFieldValue(fieldName string, value interface{}) bool {
	// Check if the value is an array
	if _, ok := value.([]interface{}); ok {
		return true
	}
	// Note: This function should be removed or made configurable
	// The array field detection should be based on schema information, not hardcoded field names
	return false
}

// processArrayField handles array fields by creating individual objects for each array element
func processArrayField(responseData map[string]interface{}, path []string, fieldName string, arrayValue interface{}, federatedResponse *FederationResponse, queryAST *ast.Document) {
	// Get the array data
	arrayData, ok := arrayValue.([]interface{})
	if !ok {
		fmt.Printf("Warning: Expected array for field %s, got %T\n", fieldName, arrayValue)
		return
	}

	// Create the destination array
	destinationArray := make([]map[string]interface{}, 0, len(arrayData))

	// Process each element in the array
	for _, element := range arrayData {
		elementMap, ok := element.(map[string]interface{})
		if !ok {
			fmt.Printf("Warning: Expected map for array element, got %T\n", element)
			continue
		}

		// Create a map for this array element
		destinationObject := make(map[string]interface{})

		// Extract individual fields from the element
		// Note: This hardcoded mapping should be replaced with schema-driven field mapping
		// For now, we'll copy all fields from the source element to maintain compatibility
		for providerField, value := range elementMap {
			// Use the provider field name as the response field name
			// In a real implementation, this should use schema-driven mapping
			destinationObject[providerField] = value
		}

		destinationArray = append(destinationArray, destinationObject)
	}

	// Push the completed array to the response
	fullPath := strings.Join(append(path, fieldName), ".")
	_, err := PushValue(responseData, fullPath, destinationArray)
	if err != nil {
		fmt.Printf("Error pushing array at path %s: %v\n", fullPath, err)
	}
}

// PushArrayValue is similar to PushValue but with enhanced array handling
func PushArrayValue(obj interface{}, path string, value interface{}) (interface{}, error) {
	keys := strings.Split(path, ".")
	return pushArrayRecursive(obj, keys, value)
}

func pushArrayRecursive(obj interface{}, keys []string, value interface{}) (interface{}, error) {
	if len(keys) == 0 {
		return value, nil
	}

	key := keys[0]

	switch curr := obj.(type) {
	case map[string]interface{}:
		child, ok := curr[key]
		if !ok {
			// If more keys → create map, else assign value
			if len(keys) > 1 {
				child = map[string]interface{}{}
			} else {
				curr[key] = value
				return curr, nil
			}
		}

		newChild, err := pushArrayRecursive(child, keys[1:], value)
		if err != nil {
			return nil, err
		}
		curr[key] = newChild
		return curr, nil

	case []interface{}:
		// For arrays: apply pushArrayRecursive to all elements
		newArr := make([]interface{}, len(curr))
		for i, elem := range curr {
			newChild, err := pushArrayRecursive(elem, keys, value)
			if err != nil {
				return nil, err
			}
			newArr[i] = newChild
		}
		return newArr, nil

	case nil:
		// Initialize a map if nil
		child := map[string]interface{}{}
		newChild, err := pushArrayRecursive(child, keys, value)
		if err != nil {
			return nil, err
		}
		return newChild, nil

	default:
		return nil, fmt.Errorf("unexpected type %T at key %q", obj, key)
	}
}

func getValueRecursive(data interface{}, keys []string) (interface{}, error) {
	if len(keys) == 0 {
		return data, nil
	}

	key := keys[0]

	switch curr := data.(type) {
	case map[string]interface{}:
		child, ok := curr[key]
		if !ok {
			return nil, fmt.Errorf("key %q not found", key)
		}
		return getValueRecursive(child, keys[1:])

	case []interface{}:
		// For arrays: apply getValueRecursive to all elements
		var results []interface{}
		for _, elem := range curr {
			childValue, err := getValueRecursive(elem, keys)
			if err != nil {
				return nil, err
			}
			results = append(results, childValue)
		}
		return results, nil

	default:
		return nil, fmt.Errorf("unexpected type %T at key %q", data, key)
	}
}
