package fetchers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

const defaultRobotsUserAgent = "goscrapling"

type FetchSafetyOptions struct {
	ObeyRobots           bool
	RobotsUserAgent      string
	BlockPrivateNetworks bool
	BlockedCIDRs         []string
}

func enforceFetchSafety(method, rawURL string, opts RequestOptions, client *http.Client) error {
	if err := blockUnsafeTarget(method, rawURL, opts.Safety); err != nil {
		return err
	}
	if opts.Safety.ObeyRobots {
		if err := checkRobotsAllowed(rawURL, opts.Safety, client); err != nil {
			return &FetcherError{Kind: FetcherErrorRobots, Method: method, URL: rawURL, Err: err}
		}
	}
	return nil
}

func blockUnsafeTarget(method, rawURL string, safety FetchSafetyOptions) error {
	if !safety.BlockPrivateNetworks && len(safety.BlockedCIDRs) == 0 {
		return nil
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	host := parsed.Hostname()
	if host == "" {
		return nil
	}

	addr, ok := parseHostAddr(host)
	if !ok {
		return nil
	}
	if safety.BlockPrivateNetworks && isPrivateAddr(addr) {
		return &FetcherError{Kind: FetcherErrorPrivateAddress, Method: method, URL: rawURL, Err: ErrPrivateAddressBlocked}
	}

	prefixes, err := parseSafetyCIDRs(safety.BlockedCIDRs)
	if err != nil {
		return err
	}
	for _, prefix := range prefixes {
		if prefix.Contains(addr) {
			return &FetcherError{Kind: FetcherErrorPrivateAddress, Method: method, URL: rawURL, Err: ErrPrivateAddressBlocked}
		}
	}
	return nil
}

func parseHostAddr(host string) (netip.Addr, bool) {
	addr, err := netip.ParseAddr(host)
	if err == nil {
		return addr.Unmap(), true
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return netip.Addr{}, false
	}
	addr, ok := netip.AddrFromSlice(ips[0])
	if !ok {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

func isPrivateAddr(addr netip.Addr) bool {
	return addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsUnspecified()
}

func parseSafetyCIDRs(values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(trimmed)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid blocked CIDR %q: %w", ErrRequestOptions, value, err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}

func checkRobotsAllowed(rawURL string, safety FetchSafetyOptions, client *http.Client) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil
	}

	robotsURL := *parsed
	robotsURL.Path = "/robots.txt"
	robotsURL.RawPath = ""
	robotsURL.RawQuery = ""
	robotsURL.Fragment = ""

	request, err := http.NewRequest(http.MethodGet, robotsURL.String(), nil)
	if err != nil {
		return err
	}
	userAgent := safety.RobotsUserAgent
	if userAgent == "" {
		userAgent = defaultRobotsUserAgent
	}
	request.Header.Set("User-Agent", userAgent)

	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	rules, err := parseRobotsRules(string(body))
	if err != nil {
		return err
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	if !rules.allowed(path) {
		return ErrRobotsBlocked
	}
	return nil
}

type robotsRules struct {
	rules []robotsRule
}

type robotsRule struct {
	allow   bool
	pattern string
}

func parseRobotsRules(content string) (robotsRules, error) {
	var output robotsRules
	active := false
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	for scanner.Scan() {
		line := scanner.Text()
		if hash := strings.IndexByte(line, '#'); hash >= 0 {
			line = line[:hash]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			active = false
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "user-agent":
			active = value == "*" || strings.EqualFold(value, defaultRobotsUserAgent)
		case "allow", "disallow":
			if !active || value == "" {
				continue
			}
			output.rules = append(output.rules, robotsRule{allow: key == "allow", pattern: value})
		}
	}
	if err := scanner.Err(); err != nil {
		return robotsRules{}, err
	}
	return output, nil
}

func (r robotsRules) allowed(path string) bool {
	bestLength := -1
	allowed := true
	for _, rule := range r.rules {
		if !strings.HasPrefix(path, rule.pattern) {
			continue
		}
		if len(rule.pattern) > bestLength || (len(rule.pattern) == bestLength && rule.allow) {
			bestLength = len(rule.pattern)
			allowed = rule.allow
		}
	}
	return allowed
}
