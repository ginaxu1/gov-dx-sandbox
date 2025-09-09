module github.com/gov-dx-sandbox/exchange/consent-engine

go 1.24.6

require (
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/exchange/config v0.0.0-00010101000000-000000000000
	github.com/gov-dx-sandbox/exchange/utils v0.0.0-00010101000000-000000000000
)

replace github.com/gov-dx-sandbox/exchange/config => ./config

replace github.com/gov-dx-sandbox/exchange/utils => ./utils
