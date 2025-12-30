# OpenTelemetry + Prometheus Example

Complete example showing how to use `rgrpc` with OpenTelemetry metrics exported to Prometheus.

## Run

```bash
go run main.go
```

Metrics will be available at `http://localhost:8080/metrics`.

## Setup Prometheus

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'my-app'
    static_configs:
      - targets: ['localhost:8080']
```

## Replace Placeholder

Before running, replace:
- `subganapathy` in the import path
- `dns:///my-service:50051` with your actual gRPC service address
- Add your generated protobuf client code

