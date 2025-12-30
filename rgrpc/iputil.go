package rgrpc

import "net"

func ipStringFromAddr(a net.Addr) string {
	if a == nil {
		return ""
	}
	// Common case
	if ta, ok := a.(*net.TCPAddr); ok {
		return ipStringFromTCPAddr(ta)
	}
	// Fallback: parse from String form
	host, _, err := net.SplitHostPort(a.String())
	if err == nil {
		return trimIPv6Brackets(host)
	}
	return ""
}

func ipStringFromTCPAddr(a *net.TCPAddr) string {
	if a == nil || a.IP == nil {
		return ""
	}
	return a.IP.String()
}

func trimIPv6Brackets(host string) string {
	// net.SplitHostPort returns host without brackets for IPv6 in most cases,
	// but being defensive doesn't hurt.
	if len(host) >= 2 && host[0] == '[' && host[len(host)-1] == ']' {
		return host[1 : len(host)-1]
	}
	return host
}
