package spiders

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

const defaultConcurrentRequests = 4

type Crawler struct {
	Sessions                    *SessionManager
	Scheduler                   *Scheduler
	DefaultCallback             Callback
	OnStart                     func(context.Context, bool) error
	OnClose                     func(context.Context, Result) error
	OnError                     func(context.Context, Request, error) error
	OnScrapedItem               func(context.Context, map[string]any) (map[string]any, error)
	AllowedDomains              []string
	RobotsTxtObey               bool
	RobotsTxtManager            *RobotsTxtManager
	RobotsUserAgent             string
	ConcurrentRequests          int
	ConcurrentRequestsPerDomain int
	DownloadDelay               time.Duration
	CheckpointDir               string
	MaxBlockedRetries           int
	IsBlocked                   BlockedCheckFunc
	RetryBlockedRequest         BlockedRetryFunc
	sleep                       func(context.Context, time.Duration) error
}

func (c Crawler) Run(ctx context.Context, start []Request) (Result, error) {
	return c.run(ctx, start, nil)
}

func (c Crawler) Stream(ctx context.Context, start []Request) (<-chan map[string]any, <-chan StreamResult) {
	items := make(chan map[string]any, 100)
	done := make(chan StreamResult, 1)
	go func() {
		defer close(items)
		result, err := c.run(ctx, start, items)
		done <- StreamResult{Result: result, Err: err}
		close(done)
	}()
	return items, done
}

func (c Crawler) run(ctx context.Context, start []Request, stream chan<- map[string]any) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.Sessions == nil {
		return Result{}, fmt.Errorf("sessions are required")
	}
	runtime := newCrawlRuntime(ctx, c)
	runtime.itemStream = stream
	return runtime.run(start)
}

func (c Crawler) processRequest(ctx context.Context, request Request, domainLimiters *crawlerDomainLimiters, sleep func(context.Context, time.Duration) error) crawlerTaskResult {
	allowed, delay, err := c.robotsRequestPolicy(ctx, request)
	if err != nil {
		return crawlerTaskResult{request: request, err: err}
	}
	if !allowed {
		return crawlerTaskResult{request: request, robotsDisallowed: true}
	}

	release, err := domainLimiters.Acquire(ctx, request.URL)
	if err != nil {
		return crawlerTaskResult{request: request, err: err}
	}
	defer release()

	if delay > 0 {
		if err := sleep(ctx, delay); err != nil {
			return crawlerTaskResult{request: request, err: err}
		}
	}

	response, err := c.Sessions.Fetch(ctx, request)
	if err != nil {
		return crawlerTaskResult{request: request, err: err}
	}

	blocked, err := c.isBlocked(ctx, response)
	if err != nil {
		return crawlerTaskResult{request: request, response: response, err: err}
	}
	if blocked {
		result := crawlerTaskResult{request: request, response: response, blocked: true}
		if request.RetryCount < effectiveMaxBlockedRetries(c.MaxBlockedRetries) {
			retry, err := c.blockedRetryRequest(ctx, request, response)
			if err != nil {
				return crawlerTaskResult{request: request, response: response, err: err}
			}
			result.retry = &retry
		}
		return result
	}

	callback := request.Callback
	if callback == nil {
		callback = c.DefaultCallback
	}
	if callback == nil {
		return crawlerTaskResult{request: request, response: response}
	}

	outputs, err := callback(ctx, response)
	if err != nil {
		return crawlerTaskResult{request: request, response: response, err: err}
	}
	return crawlerTaskResult{request: request, response: response, outputs: outputs}
}

func (c Crawler) robotsRequestPolicy(ctx context.Context, request Request) (bool, time.Duration, error) {
	delay := c.DownloadDelay
	if !c.RobotsTxtObey || c.RobotsTxtManager == nil {
		return true, delay, nil
	}
	userAgent := request.Headers.Get("User-Agent")
	if userAgent == "" {
		userAgent = c.RobotsUserAgent
	}
	if userAgent == "" {
		userAgent = "*"
	}
	allowed, err := c.RobotsTxtManager.CanFetch(ctx, request.URL, request.SID, userAgent)
	if err != nil || !allowed {
		return allowed, delay, err
	}
	directives, err := c.RobotsTxtManager.DelayDirectives(ctx, request.URL, request.SID, userAgent)
	if err != nil {
		return false, delay, err
	}
	return true, directives.EffectiveDelay(delay), nil
}

func normalizeAllowedDomains(domains []string) []string {
	if len(domains) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		domain = strings.TrimSuffix(domain, ".")
		if domain == "" {
			continue
		}
		normalized = append(normalized, domain)
	}
	return normalized
}

func isDomainAllowed(rawURL string, allowedDomains []string) bool {
	if len(allowedDomains) == 0 {
		return true
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
	if host == "" {
		return false
	}
	for _, allowed := range allowedDomains {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

func effectiveConcurrentRequests(value int) int {
	if value <= 0 {
		return defaultConcurrentRequests
	}
	return value
}

func defaultCrawlerSleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func crawlerDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSuffix(parsed.Host, "."))
}

type StreamResult struct {
	Result Result
	Err    error
}

type crawlerTaskResult struct {
	request          Request
	response         Response
	outputs          []Output
	blocked          bool
	robotsDisallowed bool
	retry            *Request
	err              error
}

type crawlerDomainLimiters struct {
	limit    int
	mu       sync.Mutex
	limiters map[string]chan struct{}
}

func newCrawlerDomainLimiters(limit int) *crawlerDomainLimiters {
	return &crawlerDomainLimiters{
		limit:    limit,
		limiters: make(map[string]chan struct{}),
	}
}

func (l *crawlerDomainLimiters) Acquire(ctx context.Context, rawURL string) (func(), error) {
	if l == nil || l.limit <= 0 {
		return func() {}, nil
	}
	domain := crawlerDomain(rawURL)

	l.mu.Lock()
	limiter := l.limiters[domain]
	if limiter == nil {
		limiter = make(chan struct{}, l.limit)
		l.limiters[domain] = limiter
	}
	l.mu.Unlock()

	select {
	case limiter <- struct{}{}:
		return func() { <-limiter }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
