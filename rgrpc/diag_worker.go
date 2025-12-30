package rgrpc

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

type diagRequest struct {
	local    string
	remote   string
	remoteIP string
}

type diagWorker struct {
	cfg Config
	reg *connRegistry
	met *metrics

	ch  chan diagRequest
	lim *rate.Limiter

	stopCh chan struct{}
}

const (
	tcpInfoRatePerSec = 4.0
	tcpInfoBurst      = 8
	tcpInfoCooldown   = 10 * time.Second
)

func newDiagWorker(cfg Config, reg *connRegistry, met *metrics, parentStop <-chan struct{}) *diagWorker {
	w := &diagWorker{
		cfg:    cfg,
		reg:    reg,
		met:    met,
		ch:     make(chan diagRequest, 512),
		stopCh: make(chan struct{}),
	}

	w.lim = rate.NewLimiter(rate.Limit(tcpInfoRatePerSec), tcpInfoBurst)

	go func() {
		select {
		case <-parentStop:
			w.stop()
		case <-w.stopCh:
		}
	}()

	go w.loop()
	return w
}

func (w *diagWorker) stop() {
	select {
	case <-w.stopCh:
	default:
		close(w.stopCh)
	}
}

func (w *diagWorker) enqueuePeriodic(ci *connInfo) {
	if ci == nil || ci.tracker == nil {
		return
	}
	req := diagRequest{
		local:    ci.local,
		remote:   ci.remote,
		remoteIP: ci.remoteIP,
	}
	select {
	case w.ch <- req:
	default:
	}
}

func (w *diagWorker) loop() {
	for {
		select {
		case <-w.stopCh:
			return
		case req := <-w.ch:
			// never block: if we can't sample now, skip
			if !w.lim.Allow() {
				continue
			}

			ci := w.reg.get(req.local, req.remote)
			if ci == nil || ci.tracker == nil {
				continue
			}

			now := time.Now()

			ci.mu.Lock()
			// cooldown
			if ci.lastSampleUnix != 0 {
				last := time.Unix(0, ci.lastSampleUnix)
				if now.Sub(last) < tcpInfoCooldown {
					ci.mu.Unlock()
					continue
				}
			}

			summary, ok := ci.tracker.Sample()
			if !ok || !summary.Available {
				ci.mu.Unlock()
				continue
			}

			var retransDelta uint32
			if summary.TotalRetrans >= ci.prevTotalRetrans {
				retransDelta = summary.TotalRetrans - ci.prevTotalRetrans
			}
			ci.prevTotalRetrans = summary.TotalRetrans
			ci.lastSampleUnix = now.UnixNano()
			ci.mu.Unlock()

			// Record TCP metrics (same bounded label scheme as call histograms: remote_ip and optional method).
			w.met.recordTCP(context.Background(), req.remoteIP, summary, retransDelta)
		}
	}
}
