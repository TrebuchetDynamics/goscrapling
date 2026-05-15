package toolbelt

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/core/storage"
	"github.com/TrebuchetDynamics/goscrapling/parser"
)

type ResponseOptions struct {
	URL         string
	StatusCode  int
	Reason      string
	Headers     http.Header
	Request     RequestMetadata
	Encoding    string
	Cookies     []*http.Cookie
	History     []*Response
	Meta        map[string]any
	CapturedXHR []*Response
	Store       storage.Store
}

type RequestMetadata struct {
	Method  string
	URL     string
	Headers http.Header
}

type Response struct {
	document    *parser.Document
	url         string
	statusCode  int
	reason      string
	headers     http.Header
	request     RequestMetadata
	body        []byte
	encoding    string
	cookies     []*http.Cookie
	history     []*Response
	meta        map[string]any
	capturedXHR []*Response
}

func NewResponse(body io.Reader, opts ResponseOptions) (*Response, error) {
	rawBody := []byte{}
	if body != nil {
		var err error
		rawBody, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	document, err := parser.Parse(bytes.NewReader(rawBody), parser.ParseOptions{URL: opts.URL, Store: opts.Store})
	if err != nil {
		return nil, err
	}

	reason := opts.Reason
	if reason == "" {
		reason = http.StatusText(opts.StatusCode)
	}

	request := cloneRequestMetadata(opts.Request)
	if request.Method == "" {
		request.Method = http.MethodGet
	}
	if request.URL == "" {
		request.URL = opts.URL
	}

	headers := opts.Headers.Clone()
	cookies := cloneCookies(opts.Cookies)
	if len(cookies) == 0 {
		cookies = cookiesFromHeaders(headers)
	}

	return &Response{
		document:    document,
		url:         opts.URL,
		statusCode:  opts.StatusCode,
		reason:      reason,
		headers:     headers,
		request:     request,
		body:        append([]byte(nil), rawBody...),
		encoding:    responseEncoding(opts.Encoding, headers),
		cookies:     cookies,
		history:     cloneResponses(opts.History),
		meta:        cloneAnyMap(opts.Meta),
		capturedXHR: cloneResponses(opts.CapturedXHR),
	}, nil
}

func (r *Response) URL() string {
	if r == nil {
		return ""
	}
	return r.url
}

func (r *Response) StatusCode() int {
	if r == nil {
		return 0
	}
	return r.statusCode
}

func (r *Response) Reason() string {
	if r == nil {
		return ""
	}
	return r.reason
}

func (r *Response) Headers() http.Header {
	if r == nil {
		return nil
	}
	return r.headers.Clone()
}

func (r *Response) Encoding() string {
	if r == nil {
		return ""
	}
	return r.encoding
}

func (r *Response) Cookies() []*http.Cookie {
	if r == nil {
		return nil
	}
	return cloneCookies(r.cookies)
}

func (r *Response) History() []*Response {
	if r == nil {
		return nil
	}
	return cloneResponses(r.history)
}

func (r *Response) Meta() map[string]any {
	if r == nil {
		return nil
	}
	return cloneAnyMap(r.meta)
}

func (r *Response) MergeMeta(extra map[string]any) map[string]any {
	merged := r.Meta()
	if merged == nil {
		merged = make(map[string]any, len(extra))
	}
	for key, value := range extra {
		merged[key] = value
	}
	return merged
}

func (r *Response) CapturedXHR() []*Response {
	if r == nil {
		return nil
	}
	return cloneResponses(r.capturedXHR)
}

func (r *Response) Request() RequestMetadata {
	if r == nil {
		return RequestMetadata{}
	}
	return cloneRequestMetadata(r.request)
}

func (r *Response) CSS(selector string) parser.Selection {
	if r == nil || r.document == nil {
		return parser.Selection{}
	}
	return r.document.CSS(selector)
}

func (r *Response) XPath(expr string) parser.Selection {
	if r == nil || r.document == nil {
		return parser.Selection{}
	}
	return r.document.XPath(expr)
}

func (r *Response) Body() []byte {
	if r == nil {
		return nil
	}
	return append([]byte(nil), r.body...)
}

func (r *Response) Text() string {
	if r == nil {
		return ""
	}
	return string(r.body)
}

func (r *Response) DecodeJSON(target any) error {
	if r == nil {
		return io.EOF
	}
	return json.Unmarshal(r.body, target)
}

func cloneRequestMetadata(request RequestMetadata) RequestMetadata {
	request.Headers = request.Headers.Clone()
	return request
}

func responseEncoding(override string, headers http.Header) string {
	override = strings.TrimSpace(override)
	if override != "" {
		return override
	}

	contentType := headers.Get("Content-Type")
	if contentType != "" {
		_, params, err := mime.ParseMediaType(contentType)
		if err == nil {
			if charset := strings.TrimSpace(params["charset"]); charset != "" {
				return strings.ToLower(charset)
			}
		}
	}
	return "utf-8"
}

func cookiesFromHeaders(headers http.Header) []*http.Cookie {
	response := http.Response{Header: headers.Clone()}
	return cloneCookies(response.Cookies())
}

func cloneCookies(cookies []*http.Cookie) []*http.Cookie {
	if len(cookies) == 0 {
		return nil
	}
	cloned := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		copied := *cookie
		cloned = append(cloned, &copied)
	}
	return cloned
}

func cloneResponses(responses []*Response) []*Response {
	if len(responses) == 0 {
		return nil
	}
	return append([]*Response(nil), responses...)
}

func cloneAnyMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
