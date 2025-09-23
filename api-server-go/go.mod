module github.com/gov-dx-sandbox/api-server-go

go 1.24.6

require (
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
	github.com/lib/pq v1.10.9
)

replace github.com/gov-dx-sandbox/exchange/consent-engine => ./shared/consent-engine

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils
