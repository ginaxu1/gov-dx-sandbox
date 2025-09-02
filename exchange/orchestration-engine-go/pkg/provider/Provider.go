package provider

// Provider struct that represents a provider information.
type Provider struct {
	ServiceUrl string `json:"serviceUrl,omitempty"`
	ServiceKey string `json:"serviceKey,omitempty"`
	ApiKey     string `json:"apiKey,omitempty"`
}
