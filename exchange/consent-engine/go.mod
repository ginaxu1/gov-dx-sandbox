module github.com/gov-dx-sandbox/exchange/consent-engine

go 1.24.6

require (
	github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring v0.0.0
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
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.53.0 // indirect
	github.com/prometheus/procfs v0.15.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/runtime v0.49.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.49.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gov-dx-sandbox/exchange/shared/config => ./shared/config

replace github.com/gov-dx-sandbox/exchange/shared/constants => ./shared/constants

replace github.com/gov-dx-sandbox/exchange/shared/utils => ./shared/utils

replace github.com/ginaxu1/gov-dx-sandbox/exchange/pkg/monitoring => ../pkg/monitoring
