package rgrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func newUnaryInterceptor(h *hooks) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		st := h.pool.Get().(*callState)
		st.reset()
		st.method = method
		st.startUnix = unixNow()

		ctx = context.WithValue(ctx, callStateKey{}, st)
		err := invoker(ctx, method, req, reply, cc, opts...)

		// finalize (invoker blocks until RPC is complete for unary)
		h.finalize(ctx, st, err)

		h.pool.Put(st)
		return err
	}
}

func newStreamInterceptor(h *hooks) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		st := h.pool.Get().(*callState)
		st.reset()
		st.method = method
		st.startUnix = unixNow()
		st.isStreaming = true // Mark as streaming RPC

		ctx = context.WithValue(ctx, callStateKey{}, st)

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			// Stream creation failed, finalize immediately
			h.finalize(ctx, st, err)
			h.pool.Put(st)
			return nil, err
		}

		// For streaming RPCs, finalization happens in stats.End (see stats.go)
		// This ensures we handle all cases correctly, including client-streaming
		// where RecvMsg returns nil on success without calling RecvMsg again.
		return stream, nil
	}
}

func (h *hooks) finalize(ctx context.Context, st *callState, callErr error) {
	start := time.Unix(0, st.startUnix)

	attempts := st.attempts.Load()

	oh := st.outHeaderUnix.Load()
	op := st.outPayloadUnix.Load()
	ip := st.inPayloadUnix.Load() // TTFB marker (first InPayload)

	var total, streamEstablish, sendStall, responseWait time.Duration

	// Determine end time: use TTFB for streaming, end-to-end for unary
	var endUnix int64
	if st.isStreaming {
		// Streaming: use TTFB (first response received) to avoid measuring idle time
		if ip > 0 {
			endUnix = ip
		} else {
			// No response received, use end time as fallback
			endUnix = st.endUnix.Load()
			if endUnix == 0 {
				endUnix = unixNow()
			}
		}
	} else {
		// Unary: use End event for true end-to-end duration
		endUnix = st.endUnix.Load()
		if endUnix == 0 {
			endUnix = unixNow()
		}
	}
	end := time.Unix(0, endUnix)
	total = end.Sub(start)

	// stream_establish: start -> outheader (folds dns/pick/connect/queue for Step 1)
	if oh > 0 {
		streamEstablish = time.Unix(0, oh).Sub(start)
	}
	// send_stall: outheader -> outpayload
	if oh > 0 && op > 0 && op >= oh {
		sendStall = time.Unix(0, op).Sub(time.Unix(0, oh))
	}
	// response_wait: outpayload -> TTFB (streaming) or end (unary)
	if op > 0 {
		if st.isStreaming && ip > 0 {
			// Streaming: measure to first response (TTFB)
			responseWait = time.Unix(0, ip).Sub(time.Unix(0, op))
		} else if endUnix > 0 && endUnix >= op {
			// Unary: measure to end (true end-to-end)
			responseWait = end.Sub(time.Unix(0, op))
		}
	}

	// Always emit call metrics. TCP metrics are sampled independently via periodic sampling.
	h.metrics.recordCall(ctx, st.method, st, total, streamEstablish, sendStall, responseWait, attempts)

	_ = callErr
}
