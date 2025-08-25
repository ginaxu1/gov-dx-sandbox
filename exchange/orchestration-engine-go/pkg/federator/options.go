package federator

import "github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go/pkg/provider"

type Options struct {
	Providers []*provider.Provider `json:"providers,omitempty"`
}
