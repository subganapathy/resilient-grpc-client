package main

import (
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
)

func main() {
	// Initialize OpenTelemetry with Prometheus exporter
	reg := prom.NewRegistry()
	exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go func() {
		log.Println("Metrics server listening on :8080/metrics")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Configure rgrpc
	cfg := rgrpc.DefaultConfig()
	cfg.EnableClientSideLB = true // Recommended for Kubernetes headless services
	cfg.MetricPrefix = "rgrpc"
	rgrpc.SetDefaultConfig(cfg)

	// Create gRPC client using rgrpc (replace with your actual service)
	cc, err := rgrpc.NewClient("dns:///my-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer cc.Close() // Automatically cleans up background workers

	// Use cc exactly like a *grpc.ClientConn
	// Example:
	// client := pb.NewMyServiceClient(cc)
	// ctx := context.Background()
	// resp, err := client.MyMethod(ctx, &pb.Request{})

	log.Println("gRPC client connected. Metrics available at http://localhost:8080/metrics")
	log.Println("Keep the process running to allow Prometheus to scrape metrics...")

	// Keep running
	select {}
}
