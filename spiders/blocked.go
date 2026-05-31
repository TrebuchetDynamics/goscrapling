package spiders

import "context"

const defaultMaxBlockedRetries = 3

var defaultBlockedStatusCodes = map[int]struct{}{
	401: {},
	403: {},
	407: {},
	429: {},
	444: {},
	500: {},
	502: {},
	503: {},
	504: {},
}

type BlockedCheckFunc func(context.Context, Response) (bool, error)

type BlockedRetryFunc func(context.Context, Request, Response) (Request, error)

func IsDefaultBlockedStatus(statusCode int) bool {
	_, ok := defaultBlockedStatusCodes[statusCode]
	return ok
}

func effectiveMaxBlockedRetries(value int) int {
	if value < 0 {
		return 0
	}
	if value == 0 {
		return defaultMaxBlockedRetries
	}
	return value
}

func (c Crawler) isBlocked(ctx context.Context, response Response) (bool, error) {
	if c.IsBlocked != nil {
		return c.IsBlocked(ctx, response)
	}
	return IsDefaultBlockedStatus(response.StatusCode()), nil
}

func (c Crawler) blockedRetryRequest(ctx context.Context, request Request, response Response) (Request, error) {
	retry := request.clone()
	retry.RetryCount++
	retry.Priority--
	retry.DontFilter = true
	clearRetryProxyMeta(&retry)
	if retry.Meta == nil {
		retry.Meta = map[string]any{}
	}
	if retry.Headers == nil {
		retry.Headers = make(map[string][]string)
	}
	if c.RetryBlockedRequest != nil {
		return c.RetryBlockedRequest(ctx, retry, response)
	}
	return retry, nil
}

func clearRetryProxyMeta(request *Request) {
	if request == nil || request.Meta == nil {
		return
	}
	delete(request.Meta, "proxy")
	delete(request.Meta, "proxies")
}
