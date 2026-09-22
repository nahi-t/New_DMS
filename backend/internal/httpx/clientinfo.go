package httpx

import (
	"net"
	"net/http"
	"strings"
)

// Trusted proxies (Render, most clouds). Adjust for your deployment.
var trustedProxies = []*net.IPNet{
	mustParseCIDR("10.0.0.0/8"),
	mustParseCIDR("172.16.0.0/12"),
	mustParseCIDR("192.168.0.0/16"),
	mustParseCIDR("127.0.0.0/8"),
	mustParseCIDR("::1/128"),
}

func mustParseCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return n
}

func isTrusted(ip net.IP) bool {
	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ClientIP returns the real client IP, falling back to RemoteAddr when the
// direct peer is not a trusted proxy.
func ClientIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	remoteIP := net.ParseIP(remoteHost)

	if remoteIP == nil || !isTrusted(remoteIP) {
		return remoteHost
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if client := strings.TrimSpace(parts[0]); client != "" {
			return client
		}
	}
	return remoteHost
}
