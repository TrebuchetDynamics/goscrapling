package spider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type Crawler struct {
	Sessions        *SessionManager
	Scheduler       *Scheduler
	DefaultCallback Callback
	AllowedDomains  []string
}

func (c Crawler) Run(ctx context.Context, start []Request) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if c.Sessions == nil {
		return Result{}, fmt.Errorf("sessions are required")
	}
	allowedDomains := normalizeAllowedDomains(c.AllowedDomains)

	scheduler := c.Scheduler
	if scheduler == nil {
		scheduler = NewScheduler(SchedulerOptions{})
	}

	result := Result{Stats: Stats{Sessions: make(map[string]int)}}
	if err := c.Sessions.Start(ctx); err != nil {
		return result, err
	}
	defer c.Sessions.Close(ctx)

	for _, request := range start {
		queued, err := scheduler.Enqueue(request)
		if err != nil {
			return result, err
		}
		if !queued {
			result.Stats.Skipped++
		}
	}

	for scheduler.Len() > 0 {
		request, ok := scheduler.Dequeue()
		if !ok {
			break
		}

		response, err := c.Sessions.Fetch(ctx, request)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.Stats.Failed++
			continue
		}
		result.Stats.Requests++
		result.Stats.Sessions[response.Request.SID]++

		callback := request.Callback
		if callback == nil {
			callback = c.DefaultCallback
		}
		if callback == nil {
			continue
		}

		outputs, err := callback(ctx, response)
		if err != nil {
			result.Errors = append(result.Errors, err)
			result.Stats.Failed++
			continue
		}
		for _, output := range outputs {
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
					return result, err
				}
				if !queued {
					result.Stats.Skipped++
				}
			}
		}
	}

	return result, nil
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
