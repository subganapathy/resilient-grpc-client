package rgrpc

import (
	"errors"
	"fmt"
	"time"
)

// Config controls the behavior of the resilient gRPC client.
type Config struct {
	// EnableClientSideLB, when true, enables client-side round-robin load balancing.
	// Recommended for Kubernetes headless services or endpoint discovery scenarios.
	// When false (default), uses pick_first (appropriate for ClusterIP/VIP services).
	EnableClientSideLB bool

	// MetricPrefix is the prefix for all emitted metric names.
	// Default: "rgrpc" (produces metrics like rgrpc_call_total_ms, rgrpc_tcp_rtt_ms, etc.)
	MetricPrefix string

	// TCPMetricsInterval controls how often to sample TCP metrics for all active connections.
	// Set to 0 to disable periodic TCP sampling.
	// Default: 5 minutes
	//
	// Note: TCP sampling is rate-limited (4 samples/sec) and has per-connection
	// cooldown (10 seconds), so under load you'll sample a rotating subset of connections.
	TCPMetricsInterval time.Duration
}

// Validate checks that the Config has valid values and returns an error if not.
func (c Config) Validate() error {
	if c.MetricPrefix == "" {
		return errors.New("MetricPrefix cannot be empty")
	}

	if c.TCPMetricsInterval < 0 {
		return fmt.Errorf("TCPMetricsInterval must be >= 0, got %v", c.TCPMetricsInterval)
	}

	return nil
}

// DefaultConfig returns a Config with sensible production defaults.
// MetricPrefix is "rgrpc", EnableClientSideLB is false, and TCPMetricsInterval is 5 minutes.
func DefaultConfig() Config {
	return Config{
		EnableClientSideLB: false,
		MetricPrefix:       "rgrpc",
		TCPMetricsInterval: 5 * time.Minute,
	}
}
