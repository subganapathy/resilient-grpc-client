package rgrpc

import (
	"net"
	"sync/atomic"
	"time"
)

type callStateKey struct{}

type callState struct {
	method string

	startUnix int64

	// set by stats handler
	outHeaderUnix  atomic.Int64
	outPayloadUnix atomic.Int64
	endUnix        atomic.Int64

	// TTFB tracking (first response received)
	inHeaderUnix  atomic.Int64 // first InHeader received
	inPayloadUnix atomic.Int64 // first InPayload received

	attempts atomic.Uint32

	remoteTCP atomic.Pointer[net.TCPAddr]
	localTCP  atomic.Pointer[net.TCPAddr]

	// cached label value (to avoid recomputing/parsing multiple times)
	remoteIP atomic.Value // string

	// isStreaming: true for streaming RPCs, false for unary
	isStreaming bool
}

func (s *callState) reset() {
	s.method = ""
	s.startUnix = 0
	s.outHeaderUnix.Store(0)
	s.outPayloadUnix.Store(0)
	s.endUnix.Store(0)
	s.inHeaderUnix.Store(0)
	s.inPayloadUnix.Store(0)
	s.attempts.Store(0)
	s.remoteTCP.Store(nil)
	s.localTCP.Store(nil)
	s.remoteIP.Store("") // ok; atomic.Value requires same concrete type after first store; we always store string
	s.isStreaming = false
}

func (s *callState) setRemoteIPOnce(ip string) {
	if ip == "" {
		return
	}
	// Store only if currently empty
	if v, ok := s.remoteIP.Load().(string); ok && v != "" {
		return
	}
	s.remoteIP.Store(ip)
}

func (s *callState) getRemoteIP() string {
	if v, ok := s.remoteIP.Load().(string); ok && v != "" {
		return v
	}
	ra := s.remoteTCP.Load()
	if ra == nil {
		return "unknown"
	}
	ip := ipStringFromTCPAddr(ra)
	if ip == "" {
		return "unknown"
	}
	s.remoteIP.Store(ip)
	return ip
}

func unixNow() int64 { return time.Now().UnixNano() }
