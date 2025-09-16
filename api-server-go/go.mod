module github.com/gov-dx-sandbox/api-server-go

go 1.24.6

require (
	github.com/gov-dx-sandbox/exchange/consent-engine v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
)

replace github.com/gov-dx-sandbox/exchange/consent-engine => ../exchange/consent-engine
replace github.com/gov-dx-sandbox/exchange/shared/utils => ../exchange/shared/utils
