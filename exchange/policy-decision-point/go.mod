module github.com/gov-dx-sandbox/exchange/policy-decision-point

go 1.24.6

require (
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/exchange/shared/utils v0.0.0
	github.com/joho/godotenv v1.5.1
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.27.0 // indirect
)

replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config

replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils
