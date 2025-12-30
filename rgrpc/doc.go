// Package rgrpc provides a drop-in replacement for grpc.NewClient with automatic
// observability. It wraps gRPC client connections and emits detailed latency
// breakdowns and TCP-level diagnostics for both unary and streaming RPC calls.
//
// # Metrics Emitted
//
// All metrics use OpenTelemetry and can be exported to Prometheus. Metrics include:
//
//   - call_total_ms: Total RPC duration (end-to-end for unary, TTFB for streaming)
//   - stream_establish_ms: Time from start to first OutHeader (includes DNS, connect, queue)
//   - send_stall_ms: Time from OutHeader to first OutPayload (flow control backpressure)
//   - response_wait_ms: Time from first OutPayload to response (TTFB for streaming, end-to-end for unary)
//   - attempts_per_call: Number of retry attempts per call
//   - tcp_rtt_ms, tcp_cwnd, tcp_retrans_delta: TCP-level diagnostics (Linux only)
//
// All metrics are labeled with method (gRPC method name) and remote_ip (backend IP).
//
// # Quick Start
//
//	import (
//	    "github.com/subganapathy/resilient-grpc-client/rgrpc"
//	    "google.golang.org/grpc"
//	)
//
//	// Configure defaults (optional)
//	cfg := rgrpc.DefaultConfig()
//	cfg.EnableClientSideLB = true // For Kubernetes headless services
//	rgrpc.SetDefaultConfig(cfg)
//
//	// Use exactly like grpc.NewClient
//	cc, err := rgrpc.NewClient("dns:///my-service:50051",
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cc.Close() // Automatically cleans up background workers
//
//	// Use cc as a normal *grpc.ClientConn
//	// client := pb.NewMyServiceClient(cc)
//
// # Streaming Semantics
//
// For unary RPCs, metrics reflect true end-to-end call duration. For streaming
// RPCs, call_total_ms represents Time To First Byte (TTFB) to avoid measuring
// idle stream time. Metrics are emitted when the stream ends (stats.End event).
//
// # Performance
//
// No syscalls on the hot path. TCP_INFO sampling happens in background workers
// at configurable intervals (default: 5 minutes).
package rgrpc
