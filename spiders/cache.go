package spiders

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/TrebuchetDynamics/goscrapling"
)

type ResponseCacheManager struct {
	dir   string
	mu    sync.Mutex
	stats ResponseCacheStats
}

type ResponseCacheStats struct {
	Hits   int
	Misses int
}

func NewResponseCacheManager(cacheDir string) *ResponseCacheManager {
	return &ResponseCacheManager{dir: cacheDir}
}

func (c *ResponseCacheManager) Get(request Request) (Response, bool, error) {
	if c == nil {
		return Response{}, false, fmt.Errorf("response cache manager is nil")
	}
	fingerprint, err := request.Fingerprint(FingerprintOptions{})
	if err != nil {
		return Response{}, false, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	rawRecord, err := os.ReadFile(c.cachePath(fingerprint))
	if err != nil {
		if os.IsNotExist(err) {
			c.stats.Misses++
			return Response{}, false, nil
		}
		return Response{}, false, err
	}

	var record responseCacheRecord
	if err := json.Unmarshal(rawRecord, &record); err != nil {
		return Response{}, false, err
	}
	body, err := base64.StdEncoding.DecodeString(record.Content)
	if err != nil {
		return Response{}, false, err
	}

	method := record.Method
	if method == "" {
		method = request.MethodOrDefault()
	}
	requestURL := record.URL
	if requestURL == "" {
		requestURL = request.URL
	}
	response, err := goscrapling.NewResponse(bytes.NewReader(body), goscrapling.ResponseOptions{
		URL:        requestURL,
		StatusCode: record.Status,
		Reason:     record.Reason,
		Headers:    record.Headers,
		Encoding:   record.Encoding,
		Cookies:    record.Cookies,
		Request: goscrapling.RequestMetadata{
			Method:  method,
			URL:     requestURL,
			Headers: record.RequestHeaders,
		},
	})
	if err != nil {
		return Response{}, false, err
	}

	c.stats.Hits++
	return Response{Response: response, Request: request.clone(), Meta: cloneMeta(request.Meta)}, true, nil
}

func (c *ResponseCacheManager) Put(request Request, response Response) error {
	if c == nil {
		return fmt.Errorf("response cache manager is nil")
	}
	if response.Response == nil {
		return fmt.Errorf("cached response is nil")
	}
	fingerprint, err := request.Fingerprint(FingerprintOptions{})
	if err != nil {
		return err
	}

	metadata := response.Response.Request()
	method := metadata.Method
	if method == "" {
		method = request.MethodOrDefault()
	}
	requestURL := metadata.URL
	if requestURL == "" {
		requestURL = request.URL
	}

	record := responseCacheRecord{
		URL:            response.URL(),
		Content:        base64.StdEncoding.EncodeToString(response.Body()),
		Status:         response.StatusCode(),
		Reason:         response.Reason(),
		Encoding:       response.Encoding(),
		Cookies:        response.Cookies(),
		Headers:        response.Headers(),
		RequestHeaders: metadata.Headers.Clone(),
		Method:         method,
	}
	if record.URL == "" {
		record.URL = requestURL
	}

	rawRecord, err := json.Marshal(record)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	path := c.cachePath(fingerprint)
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, rawRecord, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return err
	}
	return nil
}

func (c *ResponseCacheManager) Clear() error {
	if c == nil {
		return fmt.Errorf("response cache manager is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		if err := os.Remove(filepath.Join(c.dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (c *ResponseCacheManager) Stats() ResponseCacheStats {
	if c == nil {
		return ResponseCacheStats{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

func (c *ResponseCacheManager) cachePath(fingerprint string) string {
	return filepath.Join(c.dir, fingerprint+".json")
}

type responseCacheRecord struct {
	URL            string         `json:"url"`
	Content        string         `json:"content"`
	Status         int            `json:"status"`
	Reason         string         `json:"reason"`
	Encoding       string         `json:"encoding"`
	Cookies        []*http.Cookie `json:"cookies"`
	Headers        http.Header    `json:"headers"`
	RequestHeaders http.Header    `json:"request_headers"`
	Method         string         `json:"method"`
}
