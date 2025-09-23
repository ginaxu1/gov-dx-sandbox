module github.com/gov-dx-sandbox/api-server-go

go 1.24.6

require (
	github.com/gov-dx-sandbox/api-server-go/pkg/errors v0.0.0
	github.com/gov-dx-sandbox/api-server-go/shared/utils v0.0.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
)

replace github.com/gov-dx-sandbox/api-server-go/pkg/errors => ./pkg/errors

replace github.com/gov-dx-sandbox/api-server-go/shared/utils => ./shared/utils
