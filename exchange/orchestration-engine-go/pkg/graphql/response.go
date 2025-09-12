package graphql

type Response struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []interface{}          `json:"errors,omitempty"`
}
