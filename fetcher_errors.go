package goscrapling

import (
	"errors"
	"fmt"
)

var (
	ErrPrivateAddressRedirect = errors.New("private address redirect blocked")
	ErrRedirectNotAllowed     = errors.New("redirect not allowed")
	ErrRequestTimeout         = errors.New("request timeout")
	ErrProxyRequest           = errors.New("proxy request failed")
	ErrRetryExhausted         = errors.New("retry attempts exhausted")
)

type FetcherErrorKind string

const (
	FetcherErrorPrivateRedirect FetcherErrorKind = "private_redirect"
	FetcherErrorRedirect        FetcherErrorKind = "redirect"
	FetcherErrorTimeout         FetcherErrorKind = "timeout"
	FetcherErrorProxy           FetcherErrorKind = "proxy"
	FetcherErrorRetryExhausted  FetcherErrorKind = "retry_exhausted"
)

type FetcherError struct {
	Kind     FetcherErrorKind
	Method   string
	URL      string
	Attempts int
	Err      error
}

func (e *FetcherError) Error() string {
	if e == nil {
		return ""
	}

	target := e.URL
	if target == "" {
		target = "<unknown>"
	}
	if e.Attempts > 0 {
		return fmt.Sprintf("%s %s failed after %d attempts: %v", e.Method, target, e.Attempts, e.Err)
	}
	return fmt.Sprintf("%s %s failed: %v", e.Method, target, e.Err)
}

func (e *FetcherError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
