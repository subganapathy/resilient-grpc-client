[![Go Reference](https://pkg.go.dev/badge/github.com/subganapathy/resilient-grpc-client.svg)](https://pkg.go.dev/github.com/subganapathy/resilient-grpc-client/rgrpc)
[![Go Report Card](https://goreportcard.com/badge/github.com/subganapathy/resilient-grpc-client)](https://goreportcard.com/report/github.com/subganapathy/resilient-grpc-client)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# resilient-grpc-client

A drop-in wrapper for `grpc.NewClient` that adds automatic observability. Emits detailed latency breakdowns (DNS/connection, send stalls, response wait) and TCP-level diagnostics (RTT, congestion window, retransmissions) for both unary and streaming gRPC calls. Helps identify network issues, flow control problems, and backend performance bottlenecks.

## When to Use It

- **Kubernetes headless services**: Enable `EnableClientSideLB=true` for per-pod attribution
- **ClusterIP/VIP services**: Works with defaults (metrics aggregate to VIP IP)
- **Service mesh (Istio/Linkerd)**: `remote_ip` will show sidecar/VIP, not backend IP
- **Production debugging**: Correlate application latency with TCP metrics to diagnose issues

## Install

```bash
go get github.com/subganapathy/resilient-grpc-client@latest
```

## Quickstart

```go
package main

import (
    "context"
    "log"
    "net/http"

    prom "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.opentelemetry.io/otel"
    otelprom "go.opentelemetry.io/otel/exporters/prometheus"
    "go.opentelemetry.io/otel/sdk/metric"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    rgrpc "github.com/subganapathy/resilient-grpc-client/rgrpc"
    pb "your-module/path/to/proto"
)

func main() {
    // Initialize OpenTelemetry with Prometheus exporter
    reg := prom.NewRegistry()
    exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
    if err != nil {
        log.Fatal(err)
    }
    provider := metric.NewMeterProvider(metric.WithReader(exporter))
    otel.SetMeterProvider(provider)

    // Expose metrics endpoint
    http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
    go http.ListenAndServe(":8080", nil)

    // Configure rgrpc (optional - defaults work for most cases)
    cfg := rgrpc.DefaultConfig()
    cfg.EnableClientSideLB = true // Recommended for Kubernetes headless services
    rgrpc.SetDefaultConfig(cfg)

    // Use exactly like grpc.NewClient
    cc, err := rgrpc.NewClient("dns:///my-service:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer cc.Close() // Automatically cleans up background workers

    // Use cc as a normal *grpc.ClientConn
    client := pb.NewMyServiceClient(cc)
    ctx := context.Background()
    resp, err := client.MyMethod(ctx, &pb.Request{})
    // ...
}
```

See [examples/otel-prometheus-client/](examples/otel-prometheus-client/) for a complete runnable example.

**Important**: If you don't configure an OpenTelemetry MeterProvider, metrics are no-op (no metrics will be emitted). You must set up OTel with a Prometheus exporter (or another exporter) before using `rgrpc.NewClient()` to see metrics.

## What Metrics Do I Get?

All metrics use OpenTelemetry and are exported to Prometheus with underscores (e.g., `rgrpc_call_total_ms`). All call metrics include `method` and `remote_ip` labels. TCP metrics include only `remote_ip`.

| Metric Name | Type | Labels | Meaning |
|------------|------|--------|---------|
| `{prefix}_call_total_ms` | Histogram | `method`, `remote_ip` | **Unary**: End-to-end call duration. **Streaming**: Time To First Byte (TTFB). Emitted when stream ends. |
| `{prefix}_stream_establish_ms` | Histogram | `method`, `remote_ip` | Time from start to first OutHeader (includes DNS, connect, queue). |
| `{prefix}_send_stall_ms` | Histogram | `method`, `remote_ip` | Time from OutHeader to first OutPayload (flow control backpressure). |
| `{prefix}_response_wait_ms` | Histogram | `method`, `remote_ip` | **Unary**: First OutPayload to end. **Streaming**: First OutPayload to TTFB. |
| `{prefix}_attempts_per_call` | Histogram | `method`, `remote_ip` | Number of retry attempts per call. |
| `{prefix}_tcp_rtt_ms` | Histogram | `remote_ip` | TCP round-trip time (Linux only, sampled periodically). |
| `{prefix}_tcp_cwnd` | Histogram | `remote_ip` | TCP congestion window in segments (from Linux TCP_INFO snd_cwnd, ≈ cwnd*MSS bytes) (Linux only). |
| `{prefix}_tcp_retrans_delta` | Histogram | `remote_ip` | Incremental retransmissions since last sample (Linux only). |

**Note**: Streaming `call_total_ms` is emitted when the stream ends (`stats.End` event), but the value represents TTFB. If a stream never ends (leaked stream), metrics won't be emitted (expected behavior).

## Debug Playbook

### High P99 latency

1. **Check breakdown**: Compare `stream_establish_ms`, `send_stall_ms`, and `response_wait_ms`
   ```promql
   histogram_quantile(0.99, sum by (le, method) (rate(rgrpc_stream_establish_ms_bucket[5m])))
   histogram_quantile(0.99, sum by (le, method) (rate(rgrpc_send_stall_ms_bucket[5m])))
   histogram_quantile(0.99, sum by (le, method) (rate(rgrpc_response_wait_ms_bucket[5m])))
   ```

2. **High `stream_establish_ms`**: DNS/connection/LB queueing issue
   - Kubernetes headless: Check DNS resolution, pod readiness
   - ClusterIP/VIP: Check load balancer health

3. **High `send_stall_ms`**: Receiver backpressure or flow control
   - Correlate with `tcp_cwnd`: Low cwnd → network congestion
   - High cwnd but high stalls → receiver not reading fast enough

4. **High `response_wait_ms` but low TCP RTT**: Backend compute time
   ```promql
   histogram_quantile(0.99, sum by (le) (rate(rgrpc_response_wait_ms_bucket[5m]))) 
   - histogram_quantile(0.99, sum by (le) (rate(rgrpc_tcp_rtt_ms_bucket[5m])))
   ```

5. **High TCP retransmissions**: Network packet loss
   ```promql
   rate(rgrpc_tcp_retrans_delta_sum[5m]) / rate(rgrpc_tcp_retrans_delta_count[5m])
   ```

For deeper implementation details, see [IMPLEMENTATION_WALKTHROUGH.md](IMPLEMENTATION_WALKTHROUGH.md) (if present).

## Configuration

Only three options are available:

```go
cfg := rgrpc.DefaultConfig()

// Enable client-side load balancing (for Kubernetes headless services)
cfg.EnableClientSideLB = true

// Custom metric prefix (default: "rgrpc")
cfg.MetricPrefix = "myapp"

// TCP sampling interval (default: 5 minutes, set to 0 to disable)
cfg.TCPMetricsInterval = 5 * time.Minute

rgrpc.SetDefaultConfig(cfg)
```

## Performance & Overhead

Benchmarked on a typical unary RPC call path (with metrics recording enabled):

- **Allocations**: ~6 allocations per unary RPC call (~128 bytes), mostly from OpenTelemetry histogram recording
- **Latency overhead**: ~300ns per call (Apple M4, ARM64)
- **Hot path**: No syscalls. Only OpenTelemetry histogram `.Record()` calls.
- **TCP diagnostics**: Background workers sample TCP_INFO at configurable intervals (default: 5 minutes), zero hot-path impact.
- **Memory**: Bounded attribute caches (4096 entries) to avoid per-call allocations.

**Note**: Benchmark numbers vary by machine, Go version, and OTel exporter setup. Treat the numbers above as a baseline example. Run `make bench` or `go test -bench=BenchmarkUnaryAllocs -benchmem ./rgrpc` to reproduce these results on your hardware.

## Limitations

- **Label cardinality**: `remote_ip` can explode with:
  - Headless services with high pod churn
  - Large fanout clients
  - Service mesh (shows sidecar/VIP, not backend)
- **Streaming semantics**: `call_total_ms` is TTFB, not stream lifetime. Metrics emitted only when stream ends.
- **TCP diagnostics**: Linux-only (TCP_INFO syscall). Gracefully disabled on non-Linux platforms.
- **TCP sampling**: Rate-limited (4 samples/sec, 10s cooldown per connection). Under load, samples a rotating subset of connections.

## License

MIT License. See [LICENSE](LICENSE) file.
