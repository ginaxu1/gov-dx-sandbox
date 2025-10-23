module github.com/gov-dx-sandbox/audit-service

go 1.24.6

require (
	github.com/gov-dx-sandbox/shared/redis v0.0.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.3.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/gov-dx-sandbox/shared/redis => ../shared/redis
