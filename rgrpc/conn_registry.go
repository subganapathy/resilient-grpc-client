package rgrpc

import (
	"context"
	"net"
	"sync"
)

type connInfo struct {
	local    string
	remote   string
	remoteIP string

	tracker *tcpTracker // nil if unavailable

	mu sync.Mutex

	// sampling state
	lastSampleUnix   int64
	prevTotalRetrans uint32
}

type connRegistry struct {
	mu sync.RWMutex
	m  map[string]*connInfo
}

func newConnRegistry() *connRegistry {
	return &connRegistry{m: make(map[string]*connInfo)}
}

func (r *connRegistry) key(local, remote string) string {
	return local + "->" + remote
}

func (r *connRegistry) get(local, remote string) *connInfo {
	r.mu.RLock()
	ci := r.m[r.key(local, remote)]
	r.mu.RUnlock()
	return ci
}

func (r *connRegistry) snapshot() []*connInfo {
	r.mu.RLock()
	out := make([]*connInfo, 0, len(r.m))
	for _, ci := range r.m {
		out = append(out, ci)
	}
	r.mu.RUnlock()
	return out
}

func (r *connRegistry) wrapConn(ctx context.Context, c net.Conn) net.Conn {
	local := c.LocalAddr().String()
	remote := c.RemoteAddr().String()
	remoteIP := ipStringFromAddr(c.RemoteAddr())
	if remoteIP == "" {
		remoteIP = "unknown"
	}

	ci := &connInfo{
		local:    local,
		remote:   remote,
		remoteIP: remoteIP,
	}

	// TCP_INFO is supported only when the underlying conn is TCP and the platform supports it.
	if tc, ok := c.(*net.TCPConn); ok {
		if tr, err := newTCPTracker(tc); err == nil {
			ci.tracker = tr
		}
	}

	r.mu.Lock()
	r.m[r.key(local, remote)] = ci
	r.mu.Unlock()

	return &trackedConn{
		Conn: c,
		onClose: func() {
			r.mu.Lock()
			delete(r.m, r.key(local, remote))
			r.mu.Unlock()
			_ = ctx
		},
	}
}

type trackedConn struct {
	net.Conn
	onClose func()
}

func (c *trackedConn) Close() error {
	err := c.Conn.Close()
	if c.onClose != nil {
		c.onClose()
	}
	return err
}
