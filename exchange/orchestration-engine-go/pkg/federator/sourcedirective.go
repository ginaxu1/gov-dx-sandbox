package federator

import "github.com/graphql-go/graphql/language/ast"

type SourceInfo struct {
	ProviderKey   string
	ProviderField string
}

func ExtractSourceInfoFromDirective(field *ast.Field) *SourceInfo {
	if field == nil || len(field.Directives) == 0 {
		return nil
	}

	var providerKey, providerField string

	for _, dir := range field.Directives {
		if dir.Name.Value != "sourceInfo" {
			continue
		}
		for _, arg := range dir.Arguments {
			switch arg.Name.Value {
			case "providerKey":
				if strValue, ok := arg.Value.(*ast.StringValue); ok {
					providerKey = strValue.Value
				}
			case "providerField":
				if strValue, ok := arg.Value.(*ast.StringValue); ok {
					providerField = strValue.Value
				}
			}
		}
		break
	}

	if providerKey == "" && providerField == "" {
		return nil
	}

	return &SourceInfo{
		ProviderKey:   providerKey,
		ProviderField: providerField,
	}
}
