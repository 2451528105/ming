package netutil

import (
	"errors"
	"net"
)

const defaultProbeAddr = "8.8.8.8:80"

// OutboundIPv4 returns the local IPv4 used to reach external network.
// It is suitable for server runtime where network connectivity is expected.
func OutboundIPv4() (string, error) {
	return OutboundIPv4WithProbe(defaultProbeAddr)
}

// OutboundIPv4WithProbe allows custom probe target, e.g. "1.1.1.1:80".
func OutboundIPv4WithProbe(addr string) (string, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr == nil || udpAddr.IP == nil {
		return "", errors.New("failed to resolve local udp address")
	}

	ip := udpAddr.IP.To4()
	if ip == nil {
		return "", errors.New("local address is not ipv4")
	}
	return ip.String(), nil
}

// MustOutboundIPv4 returns fallback when IP detection fails.
func MustOutboundIPv4(fallback string) string {
	ip, err := OutboundIPv4()
	if err != nil {
		return fallback
	}
	return ip
}

