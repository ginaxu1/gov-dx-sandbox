module github.com/gov-dx-sandbox/exchange/consent-engine

go 1.24.6

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/exchange/shared/config v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/constants v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
	github.com/lib/pq v1.10.9
)

require (
	github.com/stretchr/testify v1.11.1
	golang.org/x/text v0.29.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config

replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils