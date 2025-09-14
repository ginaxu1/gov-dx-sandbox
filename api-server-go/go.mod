module github.com/gov-dx-sandbox/api-server-go

go 1.24.6

require (
	github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
)

replace github.com/gov-dx-sandbox/exchange/shared/utils => ../exchange/shared/utils

replace github.com/ginaxu1/gov-dx-sandbox/exchange/shared/types => ../exchange/shared/types
