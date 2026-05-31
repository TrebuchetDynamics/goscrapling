package templates

import (
	"context"

	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

type ProcessRequestFunc func(spiders.Request, spiders.Response) (spiders.Request, error)

type CrawlRule struct {
	LinkExtractor  *spiders.LinkExtractor
	Callback       spiders.Callback
	Priority       *int
	ProcessRequest ProcessRequestFunc
}

type CrawlSpider struct {
	Rules []CrawlRule
}

func (s CrawlSpider) Parse(ctx context.Context, response spiders.Response) ([]spiders.Output, error) {
	return OutputsForRules(ctx, response, s.Rules)
}

func OutputsForRules(_ context.Context, response spiders.Response, rules []CrawlRule) ([]spiders.Output, error) {
	outputs := make([]spiders.Output, 0)
	for _, rule := range rules {
		if rule.LinkExtractor == nil {
			continue
		}
		links, err := rule.LinkExtractor.Extract(response)
		if err != nil {
			return nil, err
		}
		for _, link := range links {
			request, err := response.Follow(link, spiders.FollowOptions{Callback: rule.Callback})
			if err != nil {
				return nil, err
			}
			if rule.Priority != nil {
				request.Priority = *rule.Priority
			}
			if rule.ProcessRequest != nil {
				request, err = rule.ProcessRequest(request, response)
				if err != nil {
					return nil, err
				}
			}
			outputs = append(outputs, spiders.Next(request))
		}
	}
	return outputs, nil
}

func DispatchByRules(response spiders.Response, rawURL string, rules []CrawlRule, fallbackCallback spiders.Callback) (*spiders.Request, error) {
	if len(rules) == 0 {
		request, err := response.Follow(rawURL, spiders.FollowOptions{Callback: fallbackCallback})
		if err != nil {
			return nil, err
		}
		return &request, nil
	}
	for _, rule := range rules {
		if rule.LinkExtractor == nil || !rule.LinkExtractor.Matches(rawURL) {
			continue
		}
		request, err := response.Follow(rawURL, spiders.FollowOptions{Callback: rule.Callback})
		if err != nil {
			return nil, err
		}
		if rule.Priority != nil {
			request.Priority = *rule.Priority
		}
		if rule.ProcessRequest != nil {
			request, err = rule.ProcessRequest(request, response)
			if err != nil {
				return nil, err
			}
		}
		return &request, nil
	}
	return nil, nil
}
