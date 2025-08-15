package models

type Policy struct {
	Subgraphs map[string]map[string][]string `json:"subgraphs"`
}

type RequestField struct {
	SubgraphName   string                 `json:"subgraphName"`
	TypeName       string                 `json:"typeName"`
	FieldName      string                 `json:"fieldName"`
	Classification string                 `json:"classification"`
	DLM            *string                `json:"dlm,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

type RequestBody struct {
	ConsumerID      string         `json:"consumerId"`
	RequestedFields []RequestField `json:"requestedFields"`
}

type ResponseBody struct {
	Authorized bool   `json:"authorized"`
	Message    string `json:"message"`
}
