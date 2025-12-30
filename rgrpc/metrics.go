package rgrpc

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	cfg   Config
	meter metric.Meter

	// Call histograms (ms) + attempts
	hTotal           metric.Float64Histogram
	hStreamEstablish metric.Float64Histogram
	hSendStall       metric.Float64Histogram
	hResponseWait    metric.Float64Histogram
	hAttempts        metric.Float64Histogram

	// TCP histograms
	hTCPRttMs        metric.Float64Histogram
	hTCPCwnd         metric.Float64Histogram
	hTCPRetransDelta metric.Float64Histogram

	// bounded caches of RecordOptions (avoid per-call attribute allocations)
	callCache callOptCache
	tcpCache  tcpOptCache
}

type callAttrKey struct {
	method   string
	remoteIP string
}

type callOptCache struct {
	mu  sync.Mutex
	m   map[callAttrKey]metric.RecordOption
	max int
}

type tcpOptCache struct {
	mu  sync.Mutex
	m   map[string]metric.RecordOption // key: remoteIP (or "" if label disabled)
	max int
}

const (
	maxAttrCacheSize = 4096
)

func newMetrics(cfg Config) *metrics {
	m := &metrics{cfg: cfg}
	mp := otel.GetMeterProvider()
	m.meter = mp.Meter(cfg.MetricPrefix)

	m.hTotal = mustHist(m.meter, cfg.MetricPrefix+".call_total_ms")
	m.hStreamEstablish = mustHist(m.meter, cfg.MetricPrefix+".stream_establish_ms")
	m.hSendStall = mustHist(m.meter, cfg.MetricPrefix+".send_stall_ms")
	m.hResponseWait = mustHist(m.meter, cfg.MetricPrefix+".response_wait_ms")
	m.hAttempts = mustHist(m.meter, cfg.MetricPrefix+".attempts_per_call")

	m.hTCPRttMs = mustHist(m.meter, cfg.MetricPrefix+".tcp_rtt_ms")
	m.hTCPCwnd = mustHist(m.meter, cfg.MetricPrefix+".tcp_cwnd")
	m.hTCPRetransDelta = mustHist(m.meter, cfg.MetricPrefix+".tcp_retrans_delta")

	m.callCache.m = make(map[callAttrKey]metric.RecordOption)
	m.callCache.max = maxAttrCacheSize

	m.tcpCache.m = make(map[string]metric.RecordOption)
	m.tcpCache.max = maxAttrCacheSize

	return m
}

func mustHist(m metric.Meter, name string) metric.Float64Histogram {
	h, _ := m.Float64Histogram(name)
	return h
}

func (m *metrics) recordCall(ctx context.Context, method string, st *callState,
	total, streamEstablish, sendStall, responseWait time.Duration,
	attempts uint32,
) {
	if ctx == nil {
		ctx = context.Background()
	}

	remoteIP := st.getRemoteIP()
	opt := m.callRecordOption(method, remoteIP)

	m.hTotal.Record(ctx, durMs(total), opt)
	m.hStreamEstablish.Record(ctx, durMs(streamEstablish), opt)
	m.hSendStall.Record(ctx, durMs(sendStall), opt)
	m.hResponseWait.Record(ctx, durMs(responseWait), opt)
	m.hAttempts.Record(ctx, float64(attempts), opt)
}

func (m *metrics) recordTCP(ctx context.Context, remoteIP string, tcp TCPInfoSummary, retransDelta uint32) {
	if ctx == nil {
		ctx = context.Background()
	}
	opt := m.tcpRecordOption(remoteIP)

	if tcp.Available {
		m.hTCPRttMs.Record(ctx, durMs(tcp.RTT), opt)
		m.hTCPCwnd.Record(ctx, float64(tcp.SndCwnd), opt)
		m.hTCPRetransDelta.Record(ctx, float64(retransDelta), opt)
	}
}

func (m *metrics) callRecordOption(method, remoteIP string) metric.RecordOption {
	key := callAttrKey{method: method, remoteIP: remoteIP}

	m.callCache.mu.Lock()
	defer m.callCache.mu.Unlock()

	if opt, ok := m.callCache.m[key]; ok {
		return opt
	}

	// bound cache size (simple strategy: clear when too big)
	if len(m.callCache.m) >= m.callCache.max {
		m.callCache.m = make(map[callAttrKey]metric.RecordOption)
	}

	attrs := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("remote_ip", remoteIP),
	}
	opt := metric.WithAttributes(attrs...)
	m.callCache.m[key] = opt
	return opt
}

func (m *metrics) tcpRecordOption(remoteIP string) metric.RecordOption {
	m.tcpCache.mu.Lock()
	defer m.tcpCache.mu.Unlock()

	if opt, ok := m.tcpCache.m[remoteIP]; ok {
		return opt
	}

	if len(m.tcpCache.m) >= m.tcpCache.max {
		m.tcpCache.m = make(map[string]metric.RecordOption)
	}

	attrs := []attribute.KeyValue{
		attribute.String("remote_ip", remoteIP),
	}
	opt := metric.WithAttributes(attrs...)
	m.tcpCache.m[remoteIP] = opt
	return opt
}

func durMs(d time.Duration) float64 {
	if d <= 0 {
		return 0
	}
	return float64(d) / float64(time.Millisecond)
}
