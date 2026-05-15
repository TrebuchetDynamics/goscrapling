package spider

import (
	"context"
	"errors"
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
	AllowedDomains              []string
	ConcurrentRequests          int
	ConcurrentRequestsPerDomain int
	DownloadDelay               time.Duration
	sleep                       func(context.Context, time.Duration) error
}

func (c Crawler) Run(ctx context.Context, start []Request) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.Sessions == nil {
		return Result{}, fmt.Errorf("sessions are required")
	}
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	allowedDomains := normalizeAllowedDomains(c.AllowedDomains)
	concurrentRequests := effectiveConcurrentRequests(c.ConcurrentRequests)
	domainLimiters := newCrawlerDomainLimiters(c.ConcurrentRequestsPerDomain)
	sleep := c.sleep
	if sleep == nil {
		sleep = defaultCrawlerSleep
	}

	scheduler := c.Scheduler
	if scheduler == nil {
		scheduler = NewScheduler(SchedulerOptions{})
	}

	result := Result{Stats: Stats{
		ConcurrentRequests:          concurrentRequests,
		ConcurrentRequestsPerDomain: c.ConcurrentRequestsPerDomain,
		DownloadDelay:               c.DownloadDelay,
		Sessions:                    make(map[string]int),
	}}
	if err := c.Sessions.Start(runCtx); err != nil {
		return result, err
	}
	defer c.Sessions.Close(runCtx)

	for _, request := range start {
		queued, err := scheduler.Enqueue(request)
		if err != nil {
			return result, err
		}
		if !queued {
			result.Stats.Skipped++
		}
	}

	taskResults := make(chan crawlerTaskResult)
	active := 0
	var stopErr error

	for {
		for stopErr == nil && active < concurrentRequests && scheduler.Len() > 0 {
			request, ok := scheduler.Dequeue()
			if !ok {
				break
			}
			active++
			go func() {
				taskResults <- c.processRequest(runCtx, request, domainLimiters, sleep)
			}()
		}

		if active == 0 {
			if stopErr != nil {
				return result, stopErr
			}
			if scheduler.Len() == 0 {
				return result, nil
			}
		}

		done := runCtx.Done()
		if stopErr != nil {
			done = nil
		}

		select {
		case task := <-taskResults:
			active--
			if task.err != nil {
				if runCtx.Err() != nil && errors.Is(task.err, runCtx.Err()) {
					if stopErr == nil {
						stopErr = runCtx.Err()
						cancel()
					}
					continue
				}
				result.Errors = append(result.Errors, task.err)
				result.Stats.Failed++
				continue
			}

			result.Stats.Requests++
			result.Stats.Sessions[task.response.Request.SID]++
			for _, output := range task.outputs {
				if output.Item != nil {
					result.Items = append(result.Items, cloneMeta(output.Item))
					result.Stats.Items++
				}
				if output.Request != nil {
					if !isDomainAllowed(output.Request.URL, allowedDomains) {
						result.Stats.OffsiteRequests++
						continue
					}
					queued, err := scheduler.Enqueue(*output.Request)
					if err != nil {
						if stopErr == nil {
							stopErr = err
							cancel()
						}
						continue
					}
					if !queued {
						result.Stats.Skipped++
					}
				}
			}
		case <-done:
			if stopErr == nil {
				stopErr = runCtx.Err()
				cancel()
			}
		}
	}
}

func (c Crawler) processRequest(ctx context.Context, request Request, domainLimiters *crawlerDomainLimiters, sleep func(context.Context, time.Duration) error) crawlerTaskResult {
	release, err := domainLimiters.Acquire(ctx, request.URL)
	if err != nil {
		return crawlerTaskResult{request: request, err: err}
	}
	defer release()

	if c.DownloadDelay > 0 {
		if err := sleep(ctx, c.DownloadDelay); err != nil {
			return crawlerTaskResult{request: request, err: err}
		}
	}

	response, err := c.Sessions.Fetch(ctx, request)
	if err != nil {
		return crawlerTaskResult{request: request, err: err}
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

type crawlerTaskResult struct {
	request  Request
	response Response
	outputs  []Output
	err      error
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
