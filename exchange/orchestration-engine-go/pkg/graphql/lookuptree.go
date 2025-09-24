package graphql

type ArgMapping struct {
	TargetArgName string `json:"targetArgName"`
	ProviderKey   string `json:"providerKey"`
	SourceArgPath string `json:"sourceArgPath"`
	TargetArgPath string `json:"targetArgPath"`
}
type Argument struct {
	ArgName     string        `json:"argName"`
	ArgMappings []*ArgMapping `json:"argMappings"`
}
type FieldMapping struct {
	Name      string          `json:"fieldName"`
	Arguments []*Argument     `json:"arguments"`
	SubFields []*FieldMapping `json:"subFields"`
}

type MappingAST struct {
	Mappings   []*FieldMapping `json:"mappings"`
	ArgMapping []*ArgMapping   `json:"argMappings"`
}
