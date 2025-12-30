package rgrpc

import "time"

// startTCPSampler periodically enqueues TCP_INFO sampling for all active conns.
// Metrics emitted use the same label scheme as call histograms (remote_ip).
func startTCPSampler(cfg Config, reg *connRegistry, diag *diagWorker, stopCh <-chan struct{}) {
	interval := cfg.TCPMetricsInterval
	if interval <= 0 {
		return
	}

	t := time.NewTicker(interval)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-t.C:
				conns := reg.snapshot()
				for _, ci := range conns {
					diag.enqueuePeriodic(ci)
				}
			}
		}
	}()
}
