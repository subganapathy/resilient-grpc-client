//go:build !linux

package rgrpc

import (
	"errors"
	"net"
)

type tcpTracker struct{}

func newTCPTracker(c *net.TCPConn) (*tcpTracker, error) {
	return nil, errors.New("TCP_INFO not supported on this platform")
}

func (t *tcpTracker) Sample() (TCPInfoSummary, bool) {
	return TCPInfoSummary{Available: false}, false
}
