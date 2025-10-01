package federator

import (
	"strings"

	"github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/visitor"
)

// ArgSource combines ArgMapping (Representation of Mapping Table Record) with the actual AST Argument node.
type ArgSource struct {
	*graphql.ArgMapping
	*ast.Argument
}

// FindRequiredArguments identifies which arguments are required for the provider queries based on the flattened paths
// from the main query and the argument mapping table.
func FindRequiredArguments(flattenedPaths []string, argMap []*graphql.ArgMapping) []*graphql.ArgMapping {
	var requiredArgs []*graphql.ArgMapping

	for _, path := range flattenedPaths {
		for _, arg := range argMap {
			if arg == nil {
				continue
			}

			if path == arg.TargetArgPath || strings.HasPrefix(path, arg.TargetArgPath+".") && !containsArg(requiredArgs, arg) {
				requiredArgs = append(requiredArgs, arg)
			}
		}
	}

	return requiredArgs
}

// ExtractRequiredArguments matches the actual AST arguments with the required argument mappings to produce ArgSource
// instances.
func ExtractRequiredArguments(argMap []*graphql.ArgMapping, arguments []*ast.Argument) []*ArgSource {
	requiredArgs := make([]*ArgSource, 0)

	for _, arg := range arguments {
		for _, mapping := range argMap {
			if mapping == nil {
				continue
			}

			// e.g., SourceArgPath: "personInfo-nic" -> match with "nic"
			segments := strings.Split(mapping.SourceArgPath, "-")
			lastSegment := segments[len(segments)-1]

			// If the argument name matches the last segment of SourceArgPath.
			if arg.Name.Value == lastSegment && !containsArgSource(requiredArgs, mapping) {
				requiredArgs = append(requiredArgs, &ArgSource{
					ArgMapping: mapping,
					Argument:   arg,
				})
			}
		}
	}

	return requiredArgs
}

func PushArgumentsToProviderQueryAst(args []*ArgSource, queryAst *FederationServiceAST) {

	var path = make([]string, 0)

	path = append(path, queryAst.ServiceKey)

	visitor.Visit(queryAst.QueryAst, &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			// if the node is a field, append it to path
			if field, ok := p.Node.(*ast.Field); ok {
				path = append(path, field.Name.Value)

				// now check whether the current path matches any argument's TargetArgPath
				var currentPath = strings.Join(path, ".")
				for _, arg := range args {
					// Check if the current path matches the target path exactly or is a prefix
					if arg.TargetArgPath == currentPath || strings.HasPrefix(arg.TargetArgPath, currentPath+".") {
						field.Arguments = append(field.Arguments, &ast.Argument{
							Kind: kinds.Argument,
							Name: &ast.Name{
								Kind:  kinds.Name,
								Value: arg.TargetArgName,
							},
							Value: arg.Value,
						})
					}
				}
			}
			return visitor.ActionNoChange, nil
		},
		Leave: func(p visitor.VisitFuncParams) (string, interface{}) {
			// if the node is a field, pop it from path
			if _, ok := p.Node.(*ast.Field); ok {
				if len(path) > 1 {
					path = path[:len(path)-1]
				}
			}
			return visitor.ActionNoChange, nil
		},
	}, nil)
}

func containsArg(args []*graphql.ArgMapping, target *graphql.ArgMapping) bool {
	for _, arg := range args {
		if arg.TargetArgPath == target.TargetArgPath {
			return true
		}
	}
	return false
}

// FindArrayRequiredArguments identifies which arguments are required for array fields
func FindArrayRequiredArguments(flattenedPaths []string, argMap []*graphql.ArgMapping) []*graphql.ArgMapping {
	var requiredArgs []*graphql.ArgMapping

	for _, path := range flattenedPaths {
		for _, arg := range argMap {
			if arg == nil {
				continue
			}

			// Check for array field patterns
			if path == arg.TargetArgPath || strings.HasPrefix(path, arg.TargetArgPath+".") && !containsArg(requiredArgs, arg) {
				requiredArgs = append(requiredArgs, arg)
			}
		}
	}

	return requiredArgs
}

// ExtractArrayRequiredArguments matches array field arguments with the required argument mappings
func ExtractArrayRequiredArguments(argMap []*graphql.ArgMapping, arguments []*ast.Argument) []*ArgSource {
	requiredArgs := make([]*ArgSource, 0)

	for _, arg := range arguments {
		for _, mapping := range argMap {
			if mapping == nil {
				continue
			}

			// e.g., SourceArgPath: "personInfo-nic" -> match with "nic"
			segments := strings.Split(mapping.SourceArgPath, "-")
			lastSegment := segments[len(segments)-1]

			// If the argument name matches the last segment of SourceArgPath.
			if arg.Name.Value == lastSegment && !containsArgSource(requiredArgs, mapping) {
				requiredArgs = append(requiredArgs, &ArgSource{
					ArgMapping: mapping,
					Argument:   arg,
				})
			}
		}
	}

	return requiredArgs
}

func containsArgSource(args []*ArgSource, target *graphql.ArgMapping) bool {
	for _, arg := range args {
		if arg.ArgMapping == target {
			return true
		}
	}
	return false
}
