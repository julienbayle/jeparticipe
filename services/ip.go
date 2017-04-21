package services

import (
	"net"

	"github.com/ant0ine/go-json-rest/rest"
)

// Convenient method to extract IP from request
func getIp(r *rest.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = "undefined"
	}

	if r.Request.Header.Get("X-Forwarded-For") != "" {
		ip = r.Request.Header.Get("X-Forwarded-For")
	}
	if r.Request.Header.Get("X-Real-IP") != "" {
		ip = r.Request.Header.Get("X-Real-IP")
	}

	if net.ParseIP(ip) == nil {
		ip = "undefined"
	}

	return ip
}
