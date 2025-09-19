module github.com/gov-dx-sandbox/api-server-go

go 1.24.6

require github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0

require github.com/joho/godotenv v1.5.1

replace github.com/gov-dx-sandbox/exchange/consent-engine => ../exchange/consent-engine

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils
