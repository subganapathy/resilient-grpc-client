package rgrpc

import (
	"context"
	"net"
	"sync"
)

type hooks struct {
	cfg Config

	pool sync.Pool

	metrics *metrics

	reg  *connRegistry
	diag *diagWorker

	stopCh chan struct{}
}

func newHooks(cfg Config) *hooks {
	h := &hooks{cfg: cfg}
	h.stopCh = make(chan struct{})

	h.pool.New = func() any { return &callState{} }

	h.metrics = newMetrics(cfg)
	h.reg = newConnRegistry()

	// One shared worker for TCP_INFO sampling (on-demand and periodic enqueue).
	h.diag = newDiagWorker(cfg, h.reg, h.metrics, h.stopCh)

	// Start periodic TCP metrics sampling if enabled.
	if cfg.TCPMetricsInterval > 0 {
		startTCPSampler(cfg, h.reg, h.diag, h.stopCh)
	}

	return h
}

func (h *hooks) close() {
	select {
	case <-h.stopCh:
		// already closed
	default:
		close(h.stopCh)
	}
	if h.diag != nil {
		h.diag.stop()
	}
}

func (h *hooks) dial(ctx context.Context, addr string) (net.Conn, error) {
	c, err := defaultDial(ctx, addr)
	if err != nil {
		return nil, err
	}
	return h.reg.wrapConn(ctx, c), nil
}

func defaultDial(ctx context.Context, addr string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "tcp", addr)
}
