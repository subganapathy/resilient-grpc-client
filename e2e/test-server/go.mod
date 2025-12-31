module github.com/yourusername/resilient-grpc-client/e2e/test-server

go 1.25

require (
	github.com/yourusername/resilient-grpc-client/e2e/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.78.0
)

replace github.com/yourusername/resilient-grpc-client/e2e/proto => ../proto

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
