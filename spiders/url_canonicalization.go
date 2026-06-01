package spiders

import (
	"net"
	"net/url"
	"strings"
)

func normalizeParsedURLHost(parsed *url.URL) {
	if parsed == nil || parsed.Host == "" {
		return
	}
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if port != "" {
		parsed.Host = net.JoinHostPort(host, port)
		return
	}
	if strings.Contains(host, ":") {
		parsed.Host = "[" + host + "]"
		return
	}
	parsed.Host = host
}
