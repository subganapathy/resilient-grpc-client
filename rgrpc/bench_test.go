package rgrpc

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

// BenchmarkUnaryAllocs benchmarks memory allocations per unary RPC call.
// This benchmarks the actual instrumented path by calling the interceptor,
// which allocates its own callState and handles the full lifecycle.
func BenchmarkUnaryAllocs(b *testing.B) {
	cfg := DefaultConfig()
	cfg.TCPMetricsInterval = 0 // Disable TCP sampling to measure only call path overhead
	cfg.MetricPrefix = "bench"

	h := newHooks(cfg)
	defer h.close()

	ui := newUnaryInterceptor(h)
	sh := newStatsHandler(h)
	method := "/test.Service/Method"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Call the interceptor (it will allocate its own callState from pool)
		err := ui(context.Background(), method, nil, nil, nil,
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				// Emit stats events inside the invoker using the ctx passed by interceptor
				// This ensures we hit the same callState that the interceptor allocated
				sh.HandleRPC(ctx, &stats.OutHeader{})
				sh.HandleRPC(ctx, &stats.OutPayload{SentTime: time.Now()})
				sh.HandleRPC(ctx, &stats.InPayload{RecvTime: time.Now()})
				sh.HandleRPC(ctx, &stats.End{EndTime: time.Now()})
				return nil
			},
		)

		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUnaryLatency benchmarks the time overhead per unary RPC call.
func BenchmarkUnaryLatency(b *testing.B) {
	cfg := DefaultConfig()
	cfg.TCPMetricsInterval = 0
	cfg.MetricPrefix = "bench"

	h := newHooks(cfg)
	defer h.close()

	ui := newUnaryInterceptor(h)
	sh := newStatsHandler(h)
	method := "/test.Service/Method"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := ui(context.Background(), method, nil, nil, nil,
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
				sh.HandleRPC(ctx, &stats.OutHeader{})
				sh.HandleRPC(ctx, &stats.OutPayload{SentTime: time.Now()})
				sh.HandleRPC(ctx, &stats.InPayload{RecvTime: time.Now()})
				sh.HandleRPC(ctx, &stats.End{EndTime: time.Now()})
				return nil
			},
		)

		if err != nil {
			b.Fatal(err)
		}
	}
}

