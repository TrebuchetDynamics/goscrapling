package goscrapling

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ResponseOptions struct {
	URL        string
	StatusCode int
	Reason     string
	Headers    http.Header
	Request    RequestMetadata
	Store      Store
}

type RequestMetadata struct {
	Method  string
	URL     string
	Headers http.Header
}

type Response struct {
	document   *Document
	url        string
	statusCode int
	reason     string
	headers    http.Header
	request    RequestMetadata
	body       []byte
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

	document, err := Parse(bytes.NewReader(rawBody), ParseOptions{URL: opts.URL, Store: opts.Store})
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

	return &Response{
		document:   document,
		url:        opts.URL,
		statusCode: opts.StatusCode,
		reason:     reason,
		headers:    opts.Headers.Clone(),
		request:    request,
		body:       append([]byte(nil), rawBody...),
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

func (r *Response) Request() RequestMetadata {
	if r == nil {
		return RequestMetadata{}
	}
	return cloneRequestMetadata(r.request)
}

func (r *Response) CSS(selector string) Selection {
	if r == nil || r.document == nil {
		return Selection{}
	}
	return r.document.CSS(selector)
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
