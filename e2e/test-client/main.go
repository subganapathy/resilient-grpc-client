package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/subganapathy/resilient-grpc-client/e2e/proto"
	rgrpc "github.com/subganapathy/resilient-grpc-client/rgrpc"
)

func main() {
	serverAddr := flag.String("server", "test-server:50051", "gRPC server address")
	metricsPort := flag.Int("metrics-port", 8080, "Metrics HTTP port")
	flag.Parse()

	// Initialize OpenTelemetry with Prometheus exporter
	// Create a Prometheus registry and use it with the exporter
	reg := prom.NewRegistry()
	exporter, err := prometheus.New(prometheus.WithRegisterer(reg))
	if err != nil {
		log.Fatalf("Failed to create Prometheus exporter: %v", err)
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	// Expose metrics endpoint using the Prometheus registry
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go func() {
		addr := fmt.Sprintf(":%d", *metricsPort)
		log.Printf("Metrics server listening on %s/metrics", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Configure rgrpc
	cfg := rgrpc.DefaultConfig()
	cfg.EnableClientSideLB = true
	cfg.MetricPrefix = "rgrpc"
	rgrpc.SetDefaultConfig(cfg)

	// Create gRPC client using rgrpc
	cc, err := rgrpc.NewClient(*serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer cc.Close()

	client := pb.NewEchoServiceClient(cc)
	ctx := context.Background()

	log.Println("Starting to make gRPC calls...")

	// Make unary calls
	for i := 0; i < 10; i++ {
		resp, err := client.Echo(ctx, &pb.EchoRequest{
			Message: fmt.Sprintf("unary message %d", i),
		})
		if err != nil {
			log.Printf("Echo failed: %v", err)
		} else {
			log.Printf("Echo response: %s", resp.Message)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Make streaming calls
	stream, err := client.EchoStream(ctx)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb.EchoRequest{
			Message: fmt.Sprintf("stream message %d", i),
		}); err != nil {
			log.Printf("Stream send failed: %v", err)
			break
		}

		resp, err := stream.Recv()
		if err != nil {
			log.Printf("Stream recv failed: %v", err)
			break
		}
		log.Printf("Stream response: %s", resp.Message)
		time.Sleep(100 * time.Millisecond)
	}

	if err := stream.CloseSend(); err != nil {
		log.Printf("CloseSend failed: %v", err)
	}

	log.Println("Finished making gRPC calls. Keeping client alive for metrics scraping...")

	// Keep the process alive so Prometheus can scrape metrics
	select {}
}
