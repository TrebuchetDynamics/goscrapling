package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	liveE2EUserAgentToken = "goscrapling-live-e2e"
	liveE2EUserAgentValue = "goscrapling-live-e2e/1.0 (testing; https://github.com/TrebuchetDynamics/goscrapling)"
)

type robotsDecision struct {
	allowed    bool
	reason     string
	crawlDelay time.Duration
}

type robotsGroup struct {
	agents     []string
	rules      []robotsRule
	crawlDelay time.Duration
}

type robotsRule struct {
	allow   bool
	pattern string
}

func fetchRobotsDecision(ctx context.Context, client *http.Client, rawURL string, userAgentHeader string) robotsDecision {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return robotsDecision{reason: fmt.Sprintf("invalid target URL: %v", err)}
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return robotsDecision{reason: "target URL must include scheme and host"}
	}
	robotsURL := parsed.Scheme + "://" + parsed.Host + "/robots.txt"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return robotsDecision{reason: fmt.Sprintf("create robots request: %v", err)}
	}
	request.Header.Set("User-Agent", userAgentHeader)

	response, err := client.Do(request)
	if err != nil {
		return robotsDecision{reason: fmt.Sprintf("robots.txt unavailable: %v", err)}
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return robotsDecision{reason: fmt.Sprintf("robots.txt unavailable: status %d", response.StatusCode)}
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 256*1024))
	if err != nil {
		return robotsDecision{reason: fmt.Sprintf("read robots.txt: %v", err)}
	}
	return evaluateRobots(body, parsed.EscapedPath(), liveE2EUserAgentToken)
}

func evaluateRobots(body []byte, targetPath string, userAgentToken string) robotsDecision {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return robotsDecision{reason: "empty robots.txt"}
	}
	lower := strings.ToLower(string(trimmed[:min(len(trimmed), 256)]))
	if strings.HasPrefix(lower, "<!doctype html") || strings.HasPrefix(lower, "<html") || strings.Contains(lower, "<title>") {
		return robotsDecision{reason: "not a usable robots.txt"}
	}
	if targetPath == "" {
		targetPath = "/"
	}

	groups := parseRobotsGroups(string(body))
	matching, specific := matchingRobotsGroups(groups, strings.ToLower(userAgentToken))
	if len(matching) == 0 {
		return robotsDecision{allowed: true, reason: "no matching robots rules"}
	}

	var best *robotsRule
	var crawlDelay time.Duration
	for _, group := range matching {
		if group.crawlDelay > 0 && (crawlDelay == 0 || group.crawlDelay > crawlDelay) {
			crawlDelay = group.crawlDelay
		}
		for _, rule := range group.rules {
			if rule.pattern == "" {
				continue
			}
			if robotsPatternMatches(targetPath, rule.pattern) {
				if best == nil ||
					len(rule.pattern) > len(best.pattern) ||
					(len(rule.pattern) == len(best.pattern) && rule.allow && !best.allow) {
					copyRule := rule
					best = &copyRule
				}
			}
		}
	}
	if best == nil {
		scope := "wildcard"
		if specific {
			scope = "specific"
		}
		return robotsDecision{allowed: true, reason: scope + " robots group has no matching path rule", crawlDelay: crawlDelay}
	}
	if best.allow {
		return robotsDecision{allowed: true, reason: "allowed by robots rule " + best.pattern, crawlDelay: crawlDelay}
	}
	return robotsDecision{reason: "disallowed by robots rule " + best.pattern, crawlDelay: crawlDelay}
}

func parseRobotsGroups(body string) []robotsGroup {
	var groups []robotsGroup
	var current robotsGroup
	seenDirective := false
	flush := func() {
		if len(current.agents) > 0 {
			groups = append(groups, current)
		}
		current = robotsGroup{}
		seenDirective = false
	}

	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(stripRobotsComment(line))
		if line == "" {
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
			if seenDirective {
				flush()
			}
			if value != "" {
				current.agents = append(current.agents, strings.ToLower(value))
			}
		case "allow":
			seenDirective = true
			current.rules = append(current.rules, robotsRule{allow: true, pattern: value})
		case "disallow":
			seenDirective = true
			if value != "" {
				current.rules = append(current.rules, robotsRule{pattern: value})
			}
		case "crawl-delay":
			seenDirective = true
			if seconds, err := strconv.ParseFloat(value, 64); err == nil && seconds > 0 {
				current.crawlDelay = time.Duration(seconds * float64(time.Second))
			}
		}
	}
	flush()
	return groups
}

func stripRobotsComment(line string) string {
	if before, _, ok := strings.Cut(line, "#"); ok {
		return before
	}
	return line
}

func matchingRobotsGroups(groups []robotsGroup, token string) ([]robotsGroup, bool) {
	var specific []robotsGroup
	var wildcard []robotsGroup
	for _, group := range groups {
		for _, agent := range group.agents {
			switch {
			case agent == "*":
				wildcard = append(wildcard, group)
			case agent == token || strings.Contains(token, agent):
				specific = append(specific, group)
			}
		}
	}
	if len(specific) > 0 {
		return specific, true
	}
	return wildcard, false
}

func robotsPatternMatches(path string, pattern string) bool {
	if pattern == "" {
		return false
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(pattern, "$") {
		return path == strings.TrimSuffix(pattern, "$")
	}
	if strings.Contains(pattern, "*") {
		return wildcardMatch(path, pattern)
	}
	return strings.HasPrefix(path, pattern)
}

func wildcardMatch(value string, pattern string) bool {
	parts := strings.Split(pattern, "*")
	position := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		index := strings.Index(value[position:], part)
		if index < 0 {
			return false
		}
		if i == 0 && !strings.HasPrefix(pattern, "*") && index != 0 {
			return false
		}
		position += index + len(part)
	}
	if !strings.HasSuffix(pattern, "*") && len(parts) > 0 {
		last := parts[len(parts)-1]
		return strings.HasSuffix(value, last)
	}
	return true
}
