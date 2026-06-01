package spiders

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type Callback func(ctx context.Context, response Response) ([]Output, error)

type Request struct {
	URL        string
	SID        string
	Method     string
	Body       []byte
	Headers    http.Header
	Priority   int
	DontFilter bool
	RetryCount int
	Meta       map[string]any
	Callback   Callback `json:"-"`
}

type FingerprintOptions struct {
	IncludeHeaders bool
	KeepFragments  bool
}

func (r Request) MethodOrDefault() string {
	if r.Method == "" {
		return http.MethodGet
	}
	return strings.ToUpper(r.Method)
}

func (r Request) Fingerprint(opts FingerprintOptions) (string, error) {
	canonicalURL, err := canonicalizeURL(r.URL, opts.KeepFragments)
	if err != nil {
		return "", err
	}

	data := map[string]any{
		"body":   hex.EncodeToString(r.Body),
		"method": r.MethodOrDefault(),
		"sid":    r.SID,
		"url":    canonicalURL,
	}
	if opts.IncludeHeaders {
		data["headers"] = stableHeaders(r.Headers)
	}

	body, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum(body)
	return hex.EncodeToString(sum[:]), nil
}

func (r Request) clone() Request {
	r.Body = append([]byte(nil), r.Body...)
	r.Headers = r.Headers.Clone()
	r.Meta = cloneMeta(r.Meta)
	return r
}

func canonicalizeURL(rawURL string, keepFragments bool) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	normalizeParsedURLHost(parsed)
	parsed.RawQuery = parsed.Query().Encode()
	if !keepFragments {
		parsed.Fragment = ""
	}
	return parsed.String(), nil
}

func stableHeaders(headers http.Header) [][2]string {
	normalized := normalizedHeaderValues(headers)
	keys := make([]string, 0, len(normalized))
	for key := range normalized {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	output := make([][2]string, 0, len(keys))
	for _, key := range keys {
		values := normalized[key]
		sort.Strings(values)
		for _, value := range values {
			output = append(output, [2]string{key, value})
		}
	}
	return output
}

func normalizedHeaderValues(headers http.Header) map[string][]string {
	output := make(map[string][]string, len(headers))
	for key, values := range headers {
		normalizedKey := strings.ToLower(key)
		output[normalizedKey] = append(output[normalizedKey], values...)
	}
	return output
}

func cloneMeta(meta map[string]any) map[string]any {
	if meta == nil {
		return nil
	}
	output := make(map[string]any, len(meta))
	for key, value := range meta {
		output[key] = value
	}
	return output
}
