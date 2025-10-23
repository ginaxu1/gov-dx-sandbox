module github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine-go

go 1.25.0

require (
	github.com/gov-dx-sandbox/shared/redis v0.0.0
	github.com/graphql-go/graphql v0.8.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	golang.org/x/oauth2 v0.32.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
)

replace github.com/gov-dx-sandbox/shared/redis => ../../shared/redis

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
