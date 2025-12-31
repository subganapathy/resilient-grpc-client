module github.com/yourusername/resilient-grpc-client/e2e/test-client

go 1.25

require (
	github.com/prometheus/client_golang v1.20.5
	github.com/yourusername/resilient-grpc-client v0.0.0-00010101000000-000000000000
	github.com/yourusername/resilient-grpc-client/e2e/proto v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/prometheus v0.52.0
	go.opentelemetry.io/otel/sdk/metric v1.39.0
	google.golang.org/grpc v1.78.0
)

replace github.com/yourusername/resilient-grpc-client => ../..

replace github.com/yourusername/resilient-grpc-client/e2e/proto => ../proto

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.60.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
