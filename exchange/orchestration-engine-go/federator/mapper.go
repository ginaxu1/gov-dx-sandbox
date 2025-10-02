package federator

import (
	"strconv"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/printer"
)

type SchemaCollectionResponse struct {
	ProviderFieldMap    []string
	Arguments           []*ast.Argument
	VariableDefinitions []*ast.VariableDefinition
}

func QueryBuilder(maps []string, args []*ArgSource) ([]*federationServiceRequest, error) {

	// initialize return variable
	var requests = make([]*federationServiceRequest, 0)

	var queries = BuildProviderLevelQuery(maps)

	// convert the queries into federationServiceRequest
	for _, q := range queries {
		// find the arguments to the specific provider
		var providerArgs = make([]*ArgSource, 0)

		for _, arg := range args {
			if arg == nil {
				continue
			}

			if arg.ProviderKey == q.ServiceKey {
				providerArgs = append(providerArgs, arg)
			}
		}

		PushArgumentsToProviderQueryAst(providerArgs, q)

		var query = printer.Print(q.QueryAst).(string)
		println(printer.Print(q.QueryAst).(string))

		requests = append(requests, &federationServiceRequest{
			ServiceKey: q.ServiceKey,
			GraphQLRequest: graphql.Request{
				Query:     query,
				Variables: nil,
			},
		})
	}

	return requests, nil
}

// ProviderFieldMap A function to convert the directives into a map of service key to list of fields.
func ProviderFieldMap(directives []*ast.Directive) []string {
	var fieldMap = make([]string, 0)

	for _, dir := range directives {
		if dir.Name.Value == "sourceInfo" {
			var serviceKey, fieldPath string
			for _, arg := range dir.Arguments {
				if arg.Name.Value == "providerKey" {
					if val, ok := arg.Value.(*ast.StringValue); ok {
						serviceKey = val.Value
					}
				}
				if arg.Name.Value == "providerField" {
					if val, ok := arg.Value.(*ast.StringValue); ok {
						fieldPath = val.Value
					}
				}
			}
			if serviceKey != "" && fieldPath != "" {
				fieldMap = append(fieldMap, serviceKey+"."+fieldPath)
			}
		}
	}
	return fieldMap
}

func ProviderSchemaCollector(schema *ast.Document, query *ast.Document) (*SchemaCollectionResponse, error) {
	// map of service key to list of fields

	// only query is supported not mutations or subscriptions
	if len(query.Definitions) != 1 || query.Definitions[0].(*ast.OperationDefinition).Operation != "query" {
		return nil, &graphql.JSONError{
			Message: "Only query operation is supported",
		}
	}

	// iterate through the query fields
	var selections = query.Definitions[0].(*ast.OperationDefinition).SelectionSet
	// get the query object definition from the schema
	var queryObjectDef = getQueryObjectDefinition(schema)

	if queryObjectDef == nil {
		return nil, &graphql.JSONError{
			Message: "Query object definition not found in schema",
		}
	}
	var providerDirectives, arguments = recursivelyExtractSourceSchemaInfo(selections, schema, queryObjectDef, nil, nil)

	var providerFieldMap = make([]string, 0)

	providerFieldMap = ProviderFieldMap(providerDirectives)

	// get variable definitions from the query
	var variableDefinitions = query.Definitions[0].(*ast.OperationDefinition).VariableDefinitions

	return &SchemaCollectionResponse{
		ProviderFieldMap:    providerFieldMap,
		Arguments:           arguments,
		VariableDefinitions: variableDefinitions,
	}, nil
}

// This function recursively traverses the selection set to extract @sourceInfo directives.
func recursivelyExtractSourceSchemaInfo(
	selectionSet *ast.SelectionSet,
	schema *ast.Document,
	objectDefinition *ast.ObjectDefinition,
	directives []*ast.Directive,
	arguments []*ast.Argument,
) ([]*ast.Directive, []*ast.Argument) {
	// base case
	if selectionSet == nil {
		return directives, arguments
	}

	// if directives is nil, initialize it
	if directives == nil {
		directives = make([]*ast.Directive, 0)
		arguments = make([]*ast.Argument, 0)
	}

	for _, selection := range selectionSet.Selections {
		if field, ok := selection.(*ast.Field); ok {
			// Find the field definition in the schema
			var fieldDef = FindFieldDefinitionFromFieldName(field.Name.Value, schema, objectDefinition.Name.Value)

			// Check for @sourceInfo directive
			if fieldDef != nil && len(fieldDef.Directives) > 0 {
				for _, dir := range fieldDef.Directives {
					if dir.Name.Value == "sourceInfo" {
						directives = append(directives, dir)

						// push the directive to the query ast
						if field.Directives == nil {
							field.Directives = make([]*ast.Directive, 0)
						}
						field.Directives = append(field.Directives, dir)
					}
				}
			}

			if field.Arguments != nil && len(field.Arguments) > 0 {
				arguments = append(arguments, field.Arguments...)
			}

			if selection.GetSelectionSet() != nil && len(selection.GetSelectionSet().Selections) > 0 {
				// Recursively process nested selection sets
				var nestedObjectDef *ast.ObjectDefinition
				if fieldDef != nil && fieldDef.Type != nil && fieldDef.Type.GetKind() == "Named" {
					nestedObjectDef = findTopLevelObjectDefinitionInSchema(fieldDef.Type.(*ast.Named).Name.Value, schema)
				} else if fieldDef != nil && fieldDef.Type.GetKind() == "List" {
					nestedObjectDef = findTopLevelObjectDefinitionInSchema(fieldDef.Type.(*ast.List).Type.(*ast.Named).Name.Value, schema)
				}
				if nestedObjectDef != nil {
					var selectionSet = field.GetSelectionSet()
					directives, arguments = recursivelyExtractSourceSchemaInfo(selectionSet, schema, nestedObjectDef, directives, arguments)
				}
			}
		}
	}
	return directives, arguments
}

// FindFieldDefinitionFromFieldName Helper function to find a field definition in the schema by field name and parent object name
func FindFieldDefinitionFromFieldName(fieldName string, schema *ast.Document, parentObjectName string) *ast.FieldDefinition {
	// Find the parent object definition in the schema
	parentObjectDef := findTopLevelObjectDefinitionInSchema(parentObjectName, schema)
	if parentObjectDef == nil {
		return nil
	}

	// Find the field definition within the parent object
	fieldDef := findFieldDefinitionInObject(parentObjectDef, fieldName)
	return fieldDef
}

func PushArgumentValue(arg *ast.Argument, val interface{}) {
	switch v := val.(type) {
	case string:
		arg.Value = &ast.StringValue{
			Kind:  kinds.StringValue,
			Value: v,
		}
	case int:
		arg.Value = &ast.IntValue{
			Kind:  kinds.IntValue,
			Value: string(rune(v)),
		}
	case float64:
		// inside your case float64:
		arg.Value = &ast.FloatValue{
			Kind:  kinds.FloatValue,
			Value: strconv.FormatFloat(v, 'f', -1, 64),
		}
	case bool:
		arg.Value = &ast.BooleanValue{
			Kind:  kinds.BooleanValue,
			Value: v,
		}
	}
}

// PushVariablesFromVariableDefinition A function to replace variable references in arguments with actual values from the request.
func PushVariablesFromVariableDefinition(request graphql.Request, extractedArgs []*ArgSource, variableDefinitions []*ast.VariableDefinition) {
	for _, arg := range extractedArgs {
		if arg.Argument.Value.GetKind() == "Variable" {
			varName := arg.Argument.Value.(*ast.Variable).Name.Value
			if val, exists := request.Variables[varName]; exists {
				// find the corresponding variable definition
				for _, v := range variableDefinitions {
					if v.Variable.Name.Value == varName {
						// replace the argument value with the variable value
						PushArgumentValue(arg.Argument, val)
						break
					}
				}
			}
		}
	}
}

// Helper function to find a top level object field in the schema by name
func findTopLevelObjectDefinitionInSchema(objectName string, schema *ast.Document) *ast.ObjectDefinition {
	for _, def := range schema.Definitions {
		if objDef, ok := def.(*ast.ObjectDefinition); ok {
			if objDef.Name.Value == objectName {
				return objDef
			}
		}
	}
	return nil
}

// Helper function to find a field definition in an object definition by name
func findFieldDefinitionInObject(objectDef *ast.ObjectDefinition, fieldName string) *ast.FieldDefinition {
	for _, fieldDef := range objectDef.Fields {
		if fieldDef.Name.Value == fieldName {
			return fieldDef
		}
	}
	return nil
}

func getQueryObjectDefinition(schema *ast.Document) *ast.ObjectDefinition {
	for _, def := range schema.Definitions {
		if objDef, ok := def.(*ast.ObjectDefinition); ok {
			if objDef.Name.Value == "Query" {
				return objDef
			}
		}
	}
	return nil
}
