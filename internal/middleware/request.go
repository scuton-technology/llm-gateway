package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
)

var trustProxyHeaders = strings.EqualFold(os.Getenv("LLM_GATEWAY_TRUST_PROXY_HEADERS"), "true") ||
	strings.EqualFold(os.Getenv("TRUST_PROXY_HEADERS"), "true")

// ClientIP returns the caller IP. Proxy headers are ignored unless explicitly enabled.
func ClientIP(r *http.Request) string {
	if trustProxyHeaders {
		if ip := forwardedFor(r.Header.Get("X-Forwarded-For")); ip != "" {
			return ip
		}
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
	}

	return remoteAddrIP(r.RemoteAddr)
}

// IsLoopbackRequest returns true when the immediate peer is a loopback address.
func IsLoopbackRequest(r *http.Request) bool {
	ip := net.ParseIP(remoteAddrIP(r.RemoteAddr))
	return ip != nil && ip.IsLoopback()
}

// RequestIsHTTPS detects whether the original request was served over HTTPS.
func RequestIsHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if trustProxyHeaders && strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
		return true
	}
	return false
}

func forwardedFor(value string) string {
	for _, part := range strings.Split(value, ",") {
		if ip := strings.TrimSpace(part); ip != "" {
			return ip
		}
	}
	return ""
}

func remoteAddrIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr)); err == nil {
		return host
	}
	return strings.Trim(remoteAddr, "[]")
}
