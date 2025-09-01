package utils

import (
	"net"
)

// NormalizeToIPv4 attempts to convert an IP string to IPv4 when possible.
// - Returns IPv4 if input is IPv4 or IPv6-mapped IPv4
// - Maps IPv6 loopback to 127.0.0.1
// - Returns original if not convertible
func NormalizeToIPv4(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ipStr
	}
	if ip4 := ip.To4(); ip4 != nil {
		return ip4.String()
	}
	if ip.IsLoopback() {
		return "127.0.0.1"
	}
	return ipStr
}
