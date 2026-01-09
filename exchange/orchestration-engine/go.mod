module github.com/ginaxu1/gov-dx-sandbox/exchange/orchestration-engine

go 1.25.0

require (
	github.com/graphql-go/graphql v0.8.1
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/lib/pq v1.10.9
	golang.org/x/oauth2 v0.32.0
)

require (
	github.com/go-chi/chi/v5 v5.2.3
	github.com/google/uuid v1.6.0
	github.com/gov-dx-sandbox/shared/audit v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gov-dx-sandbox/shared/audit => ../../shared/audit
