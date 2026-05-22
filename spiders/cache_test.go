package spiders_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestSpiderResponseCache(t *testing.T) {
	t.Run("stores and replays binary responses by request fingerprint", func(t *testing.T) {
		dir := t.TempDir()
		cache := spiders.NewResponseCacheManager(dir)
		request := spiders.Request{
			URL:    "https://example.com/file?b=2&a=1",
			Method: http.MethodPost,
			Body:   []byte("payload"),
			Headers: http.Header{
				"X-Trace": []string{"request-1"},
			},
		}
		body := []byte{0, 1, 2, '<', 'h', '1', '>', 'O', 'K', '<', '/', 'h', '1', '>'}
		response := newCachedSpiderResponse(t, request, body, http.StatusCreated)

		if err := cache.Put(request, response); err != nil {
			t.Fatalf("Put returned error: %v", err)
		}

		fingerprint, err := request.Fingerprint(spiders.FingerprintOptions{})
		if err != nil {
			t.Fatalf("fingerprint request: %v", err)
		}
		cacheFile := filepath.Join(dir, fingerprint+".json")
		rawFile, err := os.ReadFile(cacheFile)
		if err != nil {
			t.Fatalf("read cache file: %v", err)
		}
		var stored map[string]any
		if err := json.Unmarshal(rawFile, &stored); err != nil {
			t.Fatalf("decode cache file: %v", err)
		}
		if got, want := stored["content"], base64.StdEncoding.EncodeToString(body); got != want {
			t.Fatalf("stored content = %v, want base64 body %q", got, want)
		}

		cached, ok, err := cache.Get(request)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if !ok {
			t.Fatal("expected cache hit")
		}
		if !bytes.Equal(cached.Body(), body) {
			t.Fatalf("cached body = %v, want %v", cached.Body(), body)
		}
		if cached.URL() != request.URL {
			t.Fatalf("cached URL = %q, want %q", cached.URL(), request.URL)
		}
		if cached.StatusCode() != http.StatusCreated {
			t.Fatalf("cached status = %d, want %d", cached.StatusCode(), http.StatusCreated)
		}
		if cached.Encoding() != "latin-1" {
			t.Fatalf("cached encoding = %q, want latin-1", cached.Encoding())
		}
		if got := cached.Response.Request().Headers.Get("X-Trace"); got != "request-1" {
			t.Fatalf("cached request header = %q, want request-1", got)
		}
		if got := cached.Response.Request().Method; got != http.MethodPost {
			t.Fatalf("cached request method = %q, want POST", got)
		}
		if got := cached.Headers().Get("X-Cache-Test"); got != "stored" {
			t.Fatalf("cached response header = %q, want stored", got)
		}

		stats := cache.Stats()
		if stats.Hits != 1 || stats.Misses != 0 {
			t.Fatalf("cache stats = %#v, want 1 hit and 0 misses", stats)
		}
	})

	t.Run("separates methods and records misses", func(t *testing.T) {
		dir := t.TempDir()
		cache := spiders.NewResponseCacheManager(dir)
		getRequest := spiders.Request{URL: "https://example.com/item", Method: http.MethodGet}
		postRequest := spiders.Request{URL: "https://example.com/item", Method: http.MethodPost}

		if err := cache.Put(getRequest, newCachedSpiderResponse(t, getRequest, []byte("get"), http.StatusOK)); err != nil {
			t.Fatalf("Put GET returned error: %v", err)
		}
		if _, ok, err := cache.Get(postRequest); err != nil || ok {
			t.Fatalf("POST cache lookup ok=%v err=%v, want miss without error", ok, err)
		}
		if err := cache.Put(postRequest, newCachedSpiderResponse(t, postRequest, []byte("post"), http.StatusAccepted)); err != nil {
			t.Fatalf("Put POST returned error: %v", err)
		}

		cachedGet, ok, err := cache.Get(getRequest)
		if err != nil || !ok {
			t.Fatalf("GET cache lookup ok=%v err=%v, want hit", ok, err)
		}
		cachedPost, ok, err := cache.Get(postRequest)
		if err != nil || !ok {
			t.Fatalf("POST cache lookup ok=%v err=%v, want hit", ok, err)
		}
		if string(cachedGet.Body()) != "get" || string(cachedPost.Body()) != "post" {
			t.Fatalf("cached bodies = %q/%q, want get/post", cachedGet.Body(), cachedPost.Body())
		}

		stats := cache.Stats()
		if stats.Hits != 2 || stats.Misses != 1 {
			t.Fatalf("cache stats = %#v, want 2 hits and 1 miss", stats)
		}
	})

	t.Run("clear removes cached JSON files", func(t *testing.T) {
		dir := t.TempDir()
		cache := spiders.NewResponseCacheManager(dir)
		request := spiders.Request{URL: "https://example.com/clear", Method: http.MethodGet}
		if err := cache.Put(request, newCachedSpiderResponse(t, request, []byte("clear"), http.StatusOK)); err != nil {
			t.Fatalf("Put returned error: %v", err)
		}
		if err := cache.Clear(); err != nil {
			t.Fatalf("Clear returned error: %v", err)
		}
		matches, err := filepath.Glob(filepath.Join(dir, "*.json"))
		if err != nil {
			t.Fatalf("glob cache dir: %v", err)
		}
		if len(matches) != 0 {
			t.Fatalf("expected no cached JSON files after clear, got %v", matches)
		}
		if _, ok, err := cache.Get(request); err != nil || ok {
			t.Fatalf("cache lookup after clear ok=%v err=%v, want miss without error", ok, err)
		}
	})
}

func newCachedSpiderResponse(t *testing.T, request spiders.Request, body []byte, status int) spiders.Response {
	t.Helper()
	response, err := goscrapling.NewResponse(bytes.NewReader(body), goscrapling.ResponseOptions{
		URL:        request.URL,
		StatusCode: status,
		Reason:     http.StatusText(status),
		Headers: http.Header{
			"Content-Type": []string{"text/html; charset=latin-1"},
			"X-Cache-Test": []string{"stored"},
		},
		Request: goscrapling.RequestMetadata{
			Method:  request.MethodOrDefault(),
			URL:     request.URL,
			Headers: request.Headers,
		},
	})
	if err != nil {
		t.Fatalf("new response: %v", err)
	}
	return spiders.Response{Response: response, Request: request}
}
