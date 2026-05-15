package goscrapling

import (
	"net/http"
	"net/http/httptest"
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

func TestResponseExtendedMetadata(t *testing.T) {
	t.Run("constructed response carries cookies meta encoding history and xhr", func(t *testing.T) {
		redirect, err := NewResponse(strings.NewReader(""), ResponseOptions{
			URL:        "https://example.com/old",
			StatusCode: http.StatusFound,
			Headers: http.Header{
				"Location":     []string{"/new"},
				"Content-Type": []string{"text/html; charset=iso-8859-1"},
			},
			Request: RequestMetadata{
				Method: http.MethodGet,
				URL:    "https://example.com/old",
				Headers: http.Header{
					"Referer": []string{"https://example.com/start"},
				},
			},
		})
		if err != nil {
			t.Fatalf("NewResponse redirect: %v", err)
		}

		xhr, err := NewResponse(strings.NewReader(`{"ok":true}`), ResponseOptions{
			URL:        "https://example.com/api/products",
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"Content-Type": []string{"application/json; charset=utf-8"},
			},
		})
		if err != nil {
			t.Fatalf("NewResponse xhr: %v", err)
		}

		response, err := NewResponse(strings.NewReader(`<html><body>ok</body></html>`), ResponseOptions{
			URL:        "https://example.com/new",
			StatusCode: http.StatusOK,
			Headers: http.Header{
				"Content-Type": []string{"text/html; charset=windows-1252"},
				"Set-Cookie":   []string{"sid=abc123; Path=/; HttpOnly"},
				"X-Trace":      []string{"trace-1", "trace-2"},
			},
			Request: RequestMetadata{
				Method: http.MethodPost,
				URL:    "https://example.com/new",
				Headers: http.Header{
					"Accept":       []string{"text/html"},
					"X-Request-Id": []string{"req-1"},
				},
			},
			Meta: map[string]any{
				"proxy": "local",
				"depth": 1,
			},
			History:     []*Response{redirect},
			CapturedXHR: []*Response{xhr},
		})
		if err != nil {
			t.Fatalf("NewResponse: %v", err)
		}

		if got := response.Encoding(); got != "windows-1252" {
			t.Fatalf("expected parsed encoding windows-1252, got %q", got)
		}
		cookies := response.Cookies()
		if len(cookies) != 1 || cookies[0].Name != "sid" || cookies[0].Value != "abc123" {
			t.Fatalf("expected sid cookie from Set-Cookie header, got %#v", cookies)
		}
		history := response.History()
		if len(history) != 1 || history[0].URL() != "https://example.com/old" || history[0].StatusCode() != http.StatusFound {
			t.Fatalf("unexpected history: %#v", history)
		}
		if got := history[0].Encoding(); got != "iso-8859-1" {
			t.Fatalf("expected redirect encoding iso-8859-1, got %q", got)
		}
		meta := response.Meta()
		if meta["proxy"] != "local" || meta["depth"] != 1 {
			t.Fatalf("unexpected meta: %#v", meta)
		}
		merged := response.MergeMeta(map[string]any{"depth": 2, "source": "follow"})
		if merged["proxy"] != "local" || merged["depth"] != 2 || merged["source"] != "follow" {
			t.Fatalf("unexpected merged meta: %#v", merged)
		}
		captured := response.CapturedXHR()
		if len(captured) != 1 || captured[0].URL() != "https://example.com/api/products" {
			t.Fatalf("unexpected captured xhr: %#v", captured)
		}
		if got := response.Headers().Values("X-Trace"); len(got) != 2 || got[0] != "trace-1" || got[1] != "trace-2" {
			t.Fatalf("expected multi-value response header detail, got %#v", got)
		}
		if got := response.Request().Headers.Get("X-Request-Id"); got != "req-1" {
			t.Fatalf("expected request header detail, got %q", got)
		}

		response.Meta()["proxy"] = "changed"
		response.Cookies()[0].Value = "changed"
		if response.Meta()["proxy"] != "local" {
			t.Fatalf("expected Meta to return a copy, got %#v", response.Meta())
		}
		if response.Cookies()[0].Value != "abc123" {
			t.Fatalf("expected Cookies to return copies, got %#v", response.Cookies())
		}
	})

	t.Run("fetcher records redirect history and response cookies", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/start":
				w.Header().Set("Content-Type", "text/html; charset=iso-8859-1")
				http.SetCookie(w, &http.Cookie{Name: "hop", Value: "one"})
				http.Redirect(w, r, "/final", http.StatusFound)
			case "/final":
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				http.SetCookie(w, &http.Cookie{Name: "sid", Value: "final"})
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`<html><body><article class="final">done</article></body></html>`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		t.Cleanup(server.Close)

		response, err := (Fetcher{}).Get(server.URL+"/start", RequestOptions{})
		if err != nil {
			t.Fatalf("Fetcher.Get: %v", err)
		}

		if response.URL() != server.URL+"/final" {
			t.Fatalf("expected final URL, got %q", response.URL())
		}
		if got := response.Encoding(); got != "utf-8" {
			t.Fatalf("expected final encoding utf-8, got %q", got)
		}
		cookies := response.Cookies()
		if len(cookies) != 1 || cookies[0].Name != "sid" || cookies[0].Value != "final" {
			t.Fatalf("expected final response cookie, got %#v", cookies)
		}
		history := response.History()
		if len(history) != 1 {
			t.Fatalf("expected one redirect history response, got %d", len(history))
		}
		if history[0].URL() != server.URL+"/start" || history[0].StatusCode() != http.StatusFound {
			t.Fatalf("unexpected redirect history response: url=%q status=%d", history[0].URL(), history[0].StatusCode())
		}
		if got := history[0].Headers().Get("Location"); got != "/final" {
			t.Fatalf("expected redirect Location /final, got %q", got)
		}
		if got := history[0].Encoding(); got != "iso-8859-1" {
			t.Fatalf("expected redirect encoding iso-8859-1, got %q", got)
		}
	})
}
