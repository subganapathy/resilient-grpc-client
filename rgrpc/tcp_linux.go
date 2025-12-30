//go:build linux

package rgrpc

import (
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"net"
)

type tcpTracker struct {
	raw syscall.RawConn
}

func newTCPTracker(c *net.TCPConn) (*tcpTracker, error) {
	rc, err := c.SyscallConn()
	if err != nil {
		return nil, err
	}
	return &tcpTracker{raw: rc}, nil
}

func (t *tcpTracker) Sample() (TCPInfoSummary, bool) {
	var (
		info *unix.TCPInfo
		err  error
	)

	controlErr := t.raw.Control(func(fd uintptr) {
		info, err = unix.GetsockoptTCPInfo(int(fd), unix.IPPROTO_TCP, unix.TCP_INFO)
	})
	if controlErr != nil || err != nil || info == nil {
		return TCPInfoSummary{Available: false}, false
	}

	// Linux TCP_INFO rtt is in usec.
	rtt := time.Duration(info.Rtt) * time.Microsecond

	return TCPInfoSummary{
		Available:    true,
		RTT:          rtt,
		SndCwnd:      info.Snd_cwnd,
		TotalRetrans: info.Total_retrans,
	}, true
}
