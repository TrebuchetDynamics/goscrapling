package spider

import (
	"net/http"
	"net/url"

	"github.com/TrebuchetDynamics/goscrapling"
)

type Response struct {
	*goscrapling.Response
	Request Request
	Meta    map[string]any
}

type FollowOptions struct {
	SID       string
	Callback  Callback
	Priority  *int
	Meta      map[string]any
	Headers   http.Header
	NoReferer bool
}

func Priority(value int) *int {
	return &value
}

func (r Response) Follow(rawURL string, opts FollowOptions) (Request, error) {
	resolvedURL, err := resolveURL(r.URL(), rawURL)
	if err != nil {
		return Request{}, err
	}

	request := r.Request.clone()
	request.URL = resolvedURL
	request.DontFilter = false
	if opts.SID != "" {
		request.SID = opts.SID
	}
	if opts.Callback != nil {
		request.Callback = opts.Callback
	}
	if opts.Priority != nil {
		request.Priority = *opts.Priority
	}
	request.Meta = cloneMeta(r.Meta)
	for key, value := range opts.Meta {
		if request.Meta == nil {
			request.Meta = make(map[string]any)
		}
		request.Meta[key] = value
	}

	headers := request.Headers.Clone()
	if headers == nil {
		headers = http.Header{}
	}
	for key, values := range opts.Headers {
		headers.Del(key)
		for _, value := range values {
			headers.Add(key, value)
		}
	}
	if !opts.NoReferer {
		headers.Set("Referer", r.URL())
	}
	request.Headers = headers
	return request, nil
}

func resolveURL(baseURL, rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}
