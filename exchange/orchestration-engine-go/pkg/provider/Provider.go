package provider

// Provider struct that represents a provider information.
type Provider struct {
	ServiceUrl string `json:"providerUrl,omitempty"`
	ServiceKey string `json:"providerKey,omitempty"`
	ApiKey     string `json:"apiKey,omitempty"`
}
