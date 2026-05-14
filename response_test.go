package goscrapling

import (
	"net/http"
	"strings"
	"testing"
)

func TestResponseMetadataAndSelectorContract(t *testing.T) {
	body := `<html><body><article class="product" data-sku="sku-1"><h2>Trail Kit</h2></article></body></html>`

	response, err := NewResponse(strings.NewReader(body), ResponseOptions{
		URL:        "https://example.com/products?page=1",
		StatusCode: http.StatusCreated,
		Reason:     "Created",
		Headers: http.Header{
			"Content-Type": []string{"text/html; charset=utf-8"},
			"X-Trace":      []string{"trace-1"},
		},
		Request: RequestMetadata{
			Method: http.MethodPost,
			URL:    "https://example.com/search",
			Headers: http.Header{
				"Accept": []string{"text/html"},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewResponse returned error: %v", err)
	}

	if got := response.URL(); got != "https://example.com/products?page=1" {
		t.Fatalf("expected final response URL, got %q", got)
	}
	if got := response.StatusCode(); got != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, got)
	}
	if got := response.Reason(); got != "Created" {
		t.Fatalf("expected reason Created, got %q", got)
	}
	if got := response.Headers().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("expected content type header, got %q", got)
	}

	request := response.Request()
	if request.Method != http.MethodPost {
		t.Fatalf("expected request method POST, got %q", request.Method)
	}
	if request.URL != "https://example.com/search" {
		t.Fatalf("expected original request URL, got %q", request.URL)
	}
	if got := request.Headers.Get("Accept"); got != "text/html" {
		t.Fatalf("expected request Accept header, got %q", got)
	}

	products := response.CSS(".product")
	if products.Len() != 1 {
		t.Fatalf("expected one product, got %d", products.Len())
	}
	first, ok := products.First()
	if !ok {
		t.Fatal("expected first product")
	}
	if got := first.Text(); got != "Trail Kit" {
		t.Fatalf("expected product text, got %q", got)
	}
	if got, ok := first.Attr("data-sku"); !ok || got != "sku-1" {
		t.Fatalf("expected data-sku sku-1, got %q ok=%v", got, ok)
	}

	response.Headers().Set("X-Trace", "changed")
	request.Headers.Set("Accept", "application/json")
	if got := response.Headers().Get("X-Trace"); got != "trace-1" {
		t.Fatalf("expected response headers to be copied, got %q", got)
	}
	if got := response.Request().Headers.Get("Accept"); got != "text/html" {
		t.Fatalf("expected request headers to be copied, got %q", got)
	}
}

func TestResponseBodyHelpers(t *testing.T) {
	body := `{"name":"Trail Kit","count":2}`

	response, err := NewResponse(strings.NewReader(body), ResponseOptions{
		URL:        "https://example.com/api/products/1",
		StatusCode: http.StatusOK,
		Headers: http.Header{
			"Content-Type": []string{"application/json; charset=utf-8"},
		},
	})
	if err != nil {
		t.Fatalf("NewResponse returned error: %v", err)
	}

	if got := string(response.Body()); got != body {
		t.Fatalf("expected raw body %q, got %q", body, got)
	}
	if got := response.Text(); got != body {
		t.Fatalf("expected decoded text %q, got %q", body, got)
	}

	copiedBody := response.Body()
	copiedBody[0] = '['
	if got := string(response.Body()); got != body {
		t.Fatalf("expected body bytes to be copied, got %q", got)
	}

	var payload struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	if err := response.DecodeJSON(&payload); err != nil {
		t.Fatalf("DecodeJSON returned error: %v", err)
	}
	if payload.Name != "Trail Kit" || payload.Count != 2 {
		t.Fatalf("decoded payload = %+v", payload)
	}

	invalid, err := NewResponse(strings.NewReader(`{"name":`), ResponseOptions{
		URL:        "https://example.com/api/broken",
		StatusCode: http.StatusOK,
	})
	if err != nil {
		t.Fatalf("NewResponse returned error for invalid JSON body: %v", err)
	}

	var broken map[string]any
	if err := invalid.DecodeJSON(&broken); err == nil {
		t.Fatal("expected invalid JSON to return an error")
	}
}
