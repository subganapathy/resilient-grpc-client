package rgrpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc/stats"
)

type statsHandler struct {
	h *hooks
}

func newStatsHandler(h *hooks) stats.Handler {
	return &statsHandler{h: h}
}

func (s *statsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	// Later: DNS/pick splitting lives here + resolver wrapper.
	_ = info
	return ctx
}

func (s *statsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	st, _ := ctx.Value(callStateKey{}).(*callState)
	if st == nil {
		return
	}

	switch ev := rs.(type) {
	case *stats.OutHeader:
		now := time.Now().UnixNano()
		if st.outHeaderUnix.Load() == 0 {
			st.outHeaderUnix.Store(now)
		}
		st.attempts.Add(1)

		if ra, ok := ev.RemoteAddr.(*net.TCPAddr); ok {
			if st.remoteTCP.Load() == nil {
				st.remoteTCP.Store(ra)
			}
			st.setRemoteIPOnce(ipStringFromTCPAddr(ra))
		}
		if la, ok := ev.LocalAddr.(*net.TCPAddr); ok {
			if st.localTCP.Load() == nil {
				st.localTCP.Store(la)
			}
		}

	case *stats.OutPayload:
		t := ev.SentTime
		if t.IsZero() {
			t = time.Now()
		}
		if st.outPayloadUnix.Load() == 0 {
			st.outPayloadUnix.Store(t.UnixNano())
		}

	case *stats.InHeader:
		// Track first response header (TTFB start)
		now := time.Now().UnixNano()
		if st.inHeaderUnix.Load() == 0 {
			st.inHeaderUnix.Store(now)
		}

	case *stats.InPayload:
		// Track first response payload (TTFB)
		t := ev.RecvTime
		if t.IsZero() {
			t = time.Now()
		}
		if st.inPayloadUnix.Load() == 0 {
			st.inPayloadUnix.Store(t.UnixNano())
		}

	case *stats.End:
		t := ev.EndTime
		if t.IsZero() {
			t = time.Now()
		}
		st.endUnix.Store(t.UnixNano())

		// For streaming RPCs, finalize here when the stream ends.
		// This ensures we handle all cases correctly, including client-streaming
		// where RecvMsg returns nil on success without calling RecvMsg again.
		// For unary RPCs, the interceptor handles finalization after invoker returns.
		if st.isStreaming {
			s.h.finalize(ctx, st, nil)
			s.h.pool.Put(st)
		}
	}
}

func (s *statsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}
func (s *statsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {}

var _ stats.Handler = (*statsHandler)(nil)
