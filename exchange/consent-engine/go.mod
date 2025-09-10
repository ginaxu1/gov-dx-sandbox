module github.com/gov-dx-sandbox/exchange/consent-engine

go 1.24.6

require (
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/exchange/shared/config v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/constants v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
)

replace github.com/gov-dx-sandbox/exchange/shared/config => ../shared/config

replace github.com/gov-dx-sandbox/exchange/shared/constants => ../shared/constants

replace github.com/gov-dx-sandbox/exchange/shared/utils => ../shared/utils
