package spiders

import (
	"bufio"
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RobotsFetchFunc func(ctx context.Context, robotsURL, sid string) (Response, error)

type RobotsTxtManager struct {
	fetch RobotsFetchFunc
	mu    sync.Mutex
	cache map[string]*robotsTxtParser
}

type RobotsDelayDirectives struct {
	CrawlDelay  time.Duration
	RequestRate *RobotsRequestRate
}

type RobotsRequestRate struct {
	Requests int
	Period   time.Duration
}

func NewRobotsTxtManager(fetch RobotsFetchFunc) *RobotsTxtManager {
	return &RobotsTxtManager{fetch: fetch, cache: make(map[string]*robotsTxtParser)}
}

func (m *RobotsTxtManager) CanFetch(ctx context.Context, rawURL, sid, userAgent string) (bool, error) {
	parser, err := m.parserFor(ctx, rawURL, sid)
	if err != nil {
		return false, err
	}
	return parser.canFetch(rawURL, userAgent), nil
}

func (m *RobotsTxtManager) DelayDirectives(ctx context.Context, rawURL, sid, userAgent string) (RobotsDelayDirectives, error) {
	parser, err := m.parserFor(ctx, rawURL, sid)
	if err != nil {
		return RobotsDelayDirectives{}, err
	}
	return parser.delayDirectives(userAgent), nil
}

func (m *RobotsTxtManager) Prefetch(ctx context.Context, urls []string, sid string) error {
	seen := make(map[string]struct{}, len(urls))
	for _, rawURL := range urls {
		key, _, err := robotsCacheKeyAndURL(rawURL)
		if err != nil {
			return err
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		if _, err := m.parserFor(ctx, rawURL, sid); err != nil {
			return err
		}
	}
	return nil
}

func (d RobotsDelayDirectives) EffectiveDelay(configured time.Duration) time.Duration {
	delay := configured
	if d.CrawlDelay > delay {
		delay = d.CrawlDelay
	}
	if d.RequestRate != nil && d.RequestRate.Requests > 0 {
		rateDelay := d.RequestRate.Period / time.Duration(d.RequestRate.Requests)
		if rateDelay > delay {
			delay = rateDelay
		}
	}
	return delay
}

func (m *RobotsTxtManager) parserFor(ctx context.Context, rawURL, sid string) (*robotsTxtParser, error) {
	if m == nil {
		return parseRobotsTxt(""), nil
	}
	key, robotsURL, err := robotsCacheKeyAndURL(rawURL)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	parser := m.cache[key]
	m.mu.Unlock()
	if parser != nil {
		return parser, nil
	}

	content := ""
	if m.fetch != nil {
		response, err := m.fetch(ctx, robotsURL, sid)
		if err == nil && response.Response != nil && response.StatusCode() == http.StatusOK {
			content = response.Text()
		}
	}
	parser = parseRobotsTxt(content)

	m.mu.Lock()
	defer m.mu.Unlock()
	if cached := m.cache[key]; cached != nil {
		return cached, nil
	}
	m.cache[key] = parser
	return parser, nil
}

func robotsCacheKeyAndURL(rawURL string) (string, string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	if parsed.Host == "" {
		return "", "", &url.Error{Op: "parse", URL: rawURL, Err: errMissingRobotsHost{}}
	}
	parsed.Path = "/robots.txt"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return strings.ToLower(parsed.Host), parsed.String(), nil
}

type errMissingRobotsHost struct{}

func (errMissingRobotsHost) Error() string { return "missing host" }

type robotsTxtParser struct {
	groups []robotsGroup
}

type robotsGroup struct {
	agents      []string
	rules       []robotsRule
	crawlDelay  time.Duration
	requestRate *RobotsRequestRate
}

type robotsRule struct {
	pattern string
	allow   bool
}

func parseRobotsTxt(content string) *robotsTxtParser {
	parser := &robotsTxtParser{}
	var group robotsGroup
	hasGroup := false
	hasDirectives := false

	flush := func() {
		if !hasGroup || len(group.agents) == 0 {
			group = robotsGroup{}
			hasGroup = false
			hasDirectives = false
			return
		}
		parser.groups = append(parser.groups, group)
		group = robotsGroup{}
		hasGroup = false
		hasDirectives = false
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if before, _, ok := strings.Cut(line, "#"); ok {
			line = before
		}
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
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
			if hasGroup && hasDirectives {
				flush()
			}
			hasGroup = true
			group.agents = append(group.agents, strings.ToLower(value))
		case "allow":
			if !hasGroup {
				continue
			}
			hasDirectives = true
			if value != "" {
				group.rules = append(group.rules, robotsRule{pattern: value, allow: true})
			}
		case "disallow":
			if !hasGroup {
				continue
			}
			hasDirectives = true
			if value != "" {
				group.rules = append(group.rules, robotsRule{pattern: value})
			}
		case "crawl-delay":
			if !hasGroup {
				continue
			}
			hasDirectives = true
			if delay, ok := parseRobotsDelay(value); ok {
				group.crawlDelay = delay
			}
		case "request-rate":
			if !hasGroup {
				continue
			}
			hasDirectives = true
			if rate, ok := parseRobotsRequestRate(value); ok {
				group.requestRate = rate
			}
		}
	}
	flush()
	return parser
}

func (p *robotsTxtParser) canFetch(rawURL, userAgent string) bool {
	group := p.groupFor(userAgent)
	if group == nil {
		return true
	}
	path := robotsPath(rawURL)
	bestLength := -1
	bestAllow := true
	for _, rule := range group.rules {
		if !robotsPatternMatches(rule.pattern, path) {
			continue
		}
		length := robotsRuleLength(rule.pattern)
		if length > bestLength || (length == bestLength && rule.allow) {
			bestLength = length
			bestAllow = rule.allow
		}
	}
	if bestLength == -1 {
		return true
	}
	return bestAllow
}

func (p *robotsTxtParser) delayDirectives(userAgent string) RobotsDelayDirectives {
	group := p.groupFor(userAgent)
	if group == nil {
		return RobotsDelayDirectives{}
	}
	return RobotsDelayDirectives{CrawlDelay: group.crawlDelay, RequestRate: group.requestRate}
}

func (p *robotsTxtParser) groupFor(userAgent string) *robotsGroup {
	if p == nil {
		return nil
	}
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		ua = "*"
	}
	bestIndex := -1
	bestScore := -1
	for i, group := range p.groups {
		for _, agent := range group.agents {
			score := robotsAgentMatchScore(agent, ua)
			if score > bestScore {
				bestScore = score
				bestIndex = i
			}
		}
	}
	if bestIndex == -1 {
		return nil
	}
	return &p.groups[bestIndex]
}

func robotsAgentMatchScore(agent, userAgent string) int {
	agent = strings.ToLower(strings.TrimSpace(agent))
	if agent == "" {
		return -1
	}
	if agent == "*" {
		return 1
	}
	if strings.Contains(userAgent, agent) {
		return len(agent) + 1
	}
	return -1
}

func parseRobotsDelay(value string) (time.Duration, bool) {
	seconds, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || seconds < 0 {
		return 0, false
	}
	return time.Duration(seconds * float64(time.Second)), true
}

func parseRobotsRequestRate(value string) (*RobotsRequestRate, bool) {
	left, right, ok := strings.Cut(strings.TrimSpace(value), "/")
	if !ok {
		return nil, false
	}
	requests, err := strconv.Atoi(strings.TrimSpace(left))
	if err != nil || requests <= 0 {
		return nil, false
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(right), 64)
	if err != nil || seconds < 0 {
		return nil, false
	}
	return &RobotsRequestRate{Requests: requests, Period: time.Duration(seconds * float64(time.Second))}, true
}

func robotsPath(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return path
}

func robotsPatternMatches(pattern, path string) bool {
	if pattern == "" {
		return false
	}
	endAnchored := strings.HasSuffix(pattern, "$")
	pattern = strings.TrimSuffix(pattern, "$")
	expr := regexp.QuoteMeta(pattern)
	expr = strings.ReplaceAll(expr, "\\*", ".*")
	expr = "^" + expr
	if endAnchored {
		expr += "$"
	}
	matched, err := regexp.MatchString(expr, path)
	return err == nil && matched
}

func robotsRuleLength(pattern string) int {
	pattern = strings.TrimSuffix(pattern, "$")
	pattern = strings.ReplaceAll(pattern, "*", "")
	return len(pattern)
}
