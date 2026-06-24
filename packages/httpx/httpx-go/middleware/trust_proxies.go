package middleware

import (
	"net"
	"net/http"
	"strings"
)

// Trusted header constants matching the upstream TrustProxies.

// X-Forwarded-For
// X-Forwarded-Host
// X-Forwarded-Proto
// X-Forwarded-Port

// TrustProxies validates that proxy-related headers (X-Forwarded-*) only come
// from trusted proxy IPs. Untrusted headers are stripped.
type TrustProxies struct {
	proxies     []string
	headers     int
	trustAll    bool
	trustedIPs  map[string]struct{}
	trustedNets []*net.IPNet
}

const (
	HeaderForwardedFor = 1 << iota
	HeaderForwardedHost
	HeaderForwardedProto
	HeaderForwardedPort
	HeaderForwardedAll = HeaderForwardedFor | HeaderForwardedHost | HeaderForwardedProto | HeaderForwardedPort
)

// NewTrustProxies creates the middleware. Pass "*" as a proxy to trust all.
func NewTrustProxies(proxies []string, headers int) *TrustProxies {
	trustAll := false
	trustedIPs := make(map[string]struct{})

	var trustedNets []*net.IPNet

	for _, p := range proxies {
		if p == "*" || p == "**" {
			trustAll = true

			break
		}

		if strings.Contains(p, "/") {
			_, cidr, err := net.ParseCIDR(p)

			if err == nil {
				trustedNets = append(trustedNets, cidr)
			}
		} else {
			trustedIPs[p] = struct{}{}
		}
	}

	return &TrustProxies{
		proxies:     proxies,
		headers:     headers,
		trustAll:    trustAll,
		trustedIPs:  trustedIPs,
		trustedNets: trustedNets,
	}
}

// Wrap returns an http.Handler that strips untrusted forwarded headers.
func (m *TrustProxies) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.isTrustedProxy(r) {
			// Strip forwarded headers from untrusted proxies.
			if m.headers&HeaderForwardedFor != 0 {
				r.Header.Del("X-Forwarded-For")
			}

			if m.headers&HeaderForwardedHost != 0 {
				r.Header.Del("X-Forwarded-Host")
			}

			if m.headers&HeaderForwardedProto != 0 {
				r.Header.Del("X-Forwarded-Proto")
			}

			if m.headers&HeaderForwardedPort != 0 {
				r.Header.Del("X-Forwarded-Port")
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (m *TrustProxies) isTrustedProxy(r *http.Request) bool {
	if m.trustAll {
		return true
	}

	remoteIP := extractIP(r.RemoteAddr)
	ip := net.ParseIP(remoteIP)

	if _, ok := m.trustedIPs[remoteIP]; ok {
		return true
	}

	for _, cidr := range m.trustedNets {
		if ip != nil && cidr.Contains(ip) {
			return true
		}
	}

	return false
}

func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)

	if err != nil {
		return addr
	}

	return host
}
