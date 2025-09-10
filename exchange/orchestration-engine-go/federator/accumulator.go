package federator

import (
	"fmt"
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/federator"
	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/visitor"
)

func AccumulateResponse(queryAST *ast.Document, federatedResponse *federationResponse) graphql.Response {
	// Traverse the query AST and accumulate data from federatedResponse
	// into a single response structure.
	responseData := make(map[string]interface{})

	var path = make([]string, 0)

	visitor.Visit(queryAST, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			if node, ok := p.Node.(*ast.Field); ok {
				fieldName := node.Name.Value

				path = append(path, fieldName)

				var providerInfo = federator.ExtractSourceInfoFromDirective(node)

				if providerInfo != nil {
					var response = federatedResponse.GetProviderResponse(providerInfo.ProviderKey)

					if response != nil {
						var value, err = getValueAtPath(response.Response.Data, providerInfo.ProviderField)
						if err == nil {
							_, err = PushValue(responseData, strings.Join(path, "."), value)
							if err != nil {
								fmt.Printf("Error pushing value at path %s: %v\n", strings.Join(path, "."), err)
							}
						} else {
							fmt.Printf("Error getting value at path %s: %v\n", providerInfo.ProviderField, err)
						}
					}
				}

				//if parentNode, ok := p.Parent.(*ast.Field); ok {
				//	def := FindFieldDefinitionFromFieldName(fieldName, configs.AppConfig.Schema, parentNode.Name.Value)
				//
				//	_ = def
				//}
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

func getValueAtPath(data interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")
	return getValueRecursive(data, keys)
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
