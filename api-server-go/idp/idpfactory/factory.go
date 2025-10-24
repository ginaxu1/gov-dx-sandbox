package idpfactory

import (
	"errors"

	"github.com/gov-dx-sandbox/api-server-go/idp"
	"github.com/gov-dx-sandbox/api-server-go/idp/asgardeo"
)

type FactoryConfig struct {
	ProviderType idp.ProviderType
	BaseURL      string
	ClientID     string
	ClientSecret string
	Scopes       []string
}

func NewIdpAPIProvider(cfg FactoryConfig) (idp.IdentityProviderAPI, error) {
	switch cfg.ProviderType {
	case idp.ProviderAsgardeo:
		return asgardeo.NewClient(cfg.BaseURL, cfg.ClientID, cfg.ClientSecret, cfg.Scopes), nil
	default:
		return nil, errors.New("unsupported provider type")
	}
}
