package rgrpc

import "time"

type TCPInfoSummary struct {
	Available bool

	RTT          time.Duration
	SndCwnd      uint32
	TotalRetrans uint32
}
