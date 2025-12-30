package rgrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// ClientConn wraps grpc.ClientConn and ensures hooks are cleaned up on Close().
type ClientConn struct {
	*grpc.ClientConn
	hooks *hooks
	once  sync.Once
}

// Close closes the underlying gRPC connection and cleans up hooks (goroutines, registries).
func (c *ClientConn) Close() error {
	var err error
	c.once.Do(func() {
		if c.hooks != nil {
			c.hooks.close()
		}
		if c.ClientConn != nil {
			err = c.ClientConn.Close()
		}
	})
	return err
}

// NewClient creates a new gRPC client connection with automatic observability.
// It behaves exactly like grpc.NewClient, but adds latency breakdown metrics
// and TCP diagnostics.
//
// The connection must be closed when done to clean up background workers:
//
//	cc, err := rgrpc.NewClient("dns:///my-service:50051",
//	    grpc.WithTransportCredentials(insecure.NewCredentials()),
//	)
//	if err != nil {
//	    return err
//	}
//	defer cc.Close()
//
// Default configuration is used (see DefaultConfig and SetDefaultConfig).
func NewClient(target string, opts ...grpc.DialOption) (*ClientConn, error) {
	return NewClientWithConfig(context.Background(), target, getDefaultConfig(), opts...)
}

// NewClientWithConfig creates a new gRPC client connection with the provided configuration.
// Use this when you need custom settings (e.g., EnableClientSideLB for Kubernetes
// headless services, or custom metric prefix).
//
// The ctx parameter is reserved for future use and can be context.Background().
func NewClientWithConfig(ctx context.Context, target string, cfg Config, opts ...grpc.DialOption) (*ClientConn, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	h := newHooks(cfg)

	// Our options go first so they wrap as much as possible.
	our := []grpc.DialOption{
		grpc.WithStatsHandler(newStatsHandler(h)),
		grpc.WithChainUnaryInterceptor(newUnaryInterceptor(h)),
		grpc.WithChainStreamInterceptor(newStreamInterceptor(h)),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return h.dial(ctx, addr)
		}),
	}

	if cfg.EnableClientSideLB {
		// Round-robin is recommended for headless/endpoint-list discovery.
		our = append(our, grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`))
	}

	// NOTE: grpc.NewClient does not do I/O; it lazily connects. Preserve that.
	all := append(our, opts...)
	cc, err := grpc.NewClient(target, all...)
	if err != nil {
		h.close()
		return nil, err
	}

	_ = ctx
	return &ClientConn{
		ClientConn: cc,
		hooks:      h,
	}, nil
}
