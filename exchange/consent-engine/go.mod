module github.com/gov-dx-sandbox/exchange/consent-engine

go 1.24.6

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/exchange/shared/config v0.0.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
)

require (
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.6 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.32.0 // indirect
)

replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config

replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils
