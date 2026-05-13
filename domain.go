package goscrapling

import (
	"net"
	"net/url"
	"strings"
)

const defaultDomain = "default"

func adaptiveDomain(rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return defaultDomain
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return defaultDomain
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		host = strings.ToLower(parsed.Host)
		if splitHost, _, err := net.SplitHostPort(host); err == nil {
			host = splitHost
		}
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return defaultDomain
	}

	return host
}
