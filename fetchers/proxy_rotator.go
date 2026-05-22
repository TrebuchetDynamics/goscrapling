package fetchers

import (
	"fmt"
	"strings"
	"sync"
)

// ProxyRotationStrategy selects a proxy and the next strategy index.
// It mirrors Scrapling's callable strategy shape while keeping Go options typed.
type ProxyRotationStrategy func([]ProxyOptions, int) (ProxyOptions, int)

// ProxyRotatorOption configures a ProxyRotator.
type ProxyRotatorOption func(*proxyRotatorConfig) error

type proxyRotatorConfig struct {
	strategy ProxyRotationStrategy
}

// ProxyRotator rotates through configured proxies in a thread-safe way.
type ProxyRotator struct {
	mu           sync.Mutex
	proxies      []ProxyOptions
	strategy     ProxyRotationStrategy
	currentIndex int
}

// NewProxyRotator builds a proxy rotator from string URLs, ProxyOptions, or
// dictionary-style maps with server, username, and password keys.
func NewProxyRotator(proxies []any, options ...ProxyRotatorOption) (*ProxyRotator, error) {
	if len(proxies) == 0 {
		return nil, fmt.Errorf("%w: at least one proxy must be provided", ErrRequestOptions)
	}

	config := proxyRotatorConfig{strategy: CyclicProxyRotation}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&config); err != nil {
			return nil, err
		}
	}
	if config.strategy == nil {
		return nil, fmt.Errorf("%w: proxy rotation strategy must not be nil", ErrRequestOptions)
	}

	parsed := make([]ProxyOptions, 0, len(proxies))
	for _, proxy := range proxies {
		options, err := parseProxyRotatorEntry(proxy)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, options)
	}

	return &ProxyRotator{proxies: parsed, strategy: config.strategy}, nil
}

// WithProxyRotationStrategy configures a custom proxy rotation strategy.
func WithProxyRotationStrategy(strategy ProxyRotationStrategy) ProxyRotatorOption {
	return func(config *proxyRotatorConfig) error {
		if strategy == nil {
			return fmt.Errorf("%w: proxy rotation strategy must not be nil", ErrRequestOptions)
		}
		config.strategy = strategy
		return nil
	}
}

// CyclicProxyRotation selects proxies sequentially and wraps at the end.
func CyclicProxyRotation(proxies []ProxyOptions, currentIndex int) (ProxyOptions, int) {
	if len(proxies) == 0 {
		return ProxyOptions{}, 0
	}
	idx := currentIndex % len(proxies)
	if idx < 0 {
		idx += len(proxies)
	}
	return cloneProxyOptions(proxies[idx]), (idx + 1) % len(proxies)
}

// Next returns the next proxy selected by the rotator's strategy.
func (r *ProxyRotator) Next() (ProxyOptions, error) {
	if r == nil {
		return ProxyOptions{}, fmt.Errorf("%w: proxy rotator is nil", ErrRequestOptions)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.proxies) == 0 {
		return ProxyOptions{}, fmt.Errorf("%w: proxy rotator has no proxies", ErrRequestOptions)
	}
	if r.strategy == nil {
		return ProxyOptions{}, fmt.Errorf("%w: proxy rotation strategy must not be nil", ErrRequestOptions)
	}

	proxy, nextIndex := r.strategy(cloneProxyOptionsSlice(r.proxies), r.currentIndex)
	if nextIndex < 0 {
		return ProxyOptions{}, fmt.Errorf("%w: proxy rotation strategy returned negative index", ErrRequestOptions)
	}
	if _, ok, err := newProxyConfig(proxy); err != nil {
		return ProxyOptions{}, err
	} else if !ok {
		return ProxyOptions{}, fmt.Errorf("%w: proxy rotation strategy returned empty proxy", ErrRequestOptions)
	}

	r.currentIndex = nextIndex
	return cloneProxyOptions(proxy), nil
}

// Proxies returns a copy of all configured proxies.
func (r *ProxyRotator) Proxies() []ProxyOptions {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return cloneProxyOptionsSlice(r.proxies)
}

// Len returns the number of configured proxies.
func (r *ProxyRotator) Len() int {
	if r == nil {
		return 0
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.proxies)
}

func parseProxyRotatorEntry(entry any) (ProxyOptions, error) {
	switch proxy := entry.(type) {
	case string:
		return validateRotatedProxyOptions(ProxyOptions{URL: proxy})
	case ProxyOptions:
		return validateRotatedProxyOptions(proxy)
	case map[string]string:
		return proxyOptionsFromStringMap(proxy)
	case map[string]any:
		return proxyOptionsFromAnyMap(proxy)
	default:
		return ProxyOptions{}, fmt.Errorf("%w: invalid proxy type %T", ErrRequestOptions, entry)
	}
}

func proxyOptionsFromStringMap(values map[string]string) (ProxyOptions, error) {
	lookup := make(map[string]string, len(values))
	for key, value := range values {
		lookup[strings.ToLower(strings.TrimSpace(key))] = value
	}
	return proxyOptionsFromMapValues(lookup["server"], lookup["username"], lookup["password"])
}

func proxyOptionsFromAnyMap(values map[string]any) (ProxyOptions, error) {
	lookup := make(map[string]string, len(values))
	for key, value := range values {
		text, ok := value.(string)
		if !ok {
			return ProxyOptions{}, fmt.Errorf("%w: proxy map value %q must be a string", ErrRequestOptions, key)
		}
		lookup[strings.ToLower(strings.TrimSpace(key))] = text
	}
	return proxyOptionsFromMapValues(lookup["server"], lookup["username"], lookup["password"])
}

func proxyOptionsFromMapValues(server, username, password string) (ProxyOptions, error) {
	if strings.TrimSpace(server) == "" {
		return ProxyOptions{}, fmt.Errorf("%w: proxy dict must have a server key", ErrRequestOptions)
	}

	options := ProxyOptions{URL: server}
	if username != "" || password != "" {
		options.Auth = &BasicAuth{Username: username, Password: password}
	}
	return validateRotatedProxyOptions(options)
}

func validateRotatedProxyOptions(options ProxyOptions) (ProxyOptions, error) {
	if _, ok, err := newProxyConfig(options); err != nil {
		return ProxyOptions{}, err
	} else if !ok {
		return ProxyOptions{}, fmt.Errorf("%w: proxy must not be empty", ErrRequestOptions)
	}
	return cloneProxyOptions(options), nil
}

func cloneProxyOptionsSlice(options []ProxyOptions) []ProxyOptions {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]ProxyOptions, len(options))
	for i, option := range options {
		cloned[i] = cloneProxyOptions(option)
	}
	return cloned
}
