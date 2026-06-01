package spiders_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
	spidertemplates "github.com/TrebuchetDynamics/goscrapling/spiders/templates"
)

func TestSpiderLinkExtractorAndTemplates(t *testing.T) {
	page := templateFixture(t, "page.html")
	response := newTemplateResponse(t, "https://example.com/base/index.html", page, "text/html")

	t.Run("link extractor filters scopes canonicalizes and processes URLs", func(t *testing.T) {
		extractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Allow:        []string{`/keep`, `/post/`, `/area`, `/asset`},
			Deny:         []string{`logout`},
			AllowDomains: []string{"example.com"},
			DenyDomains:  []string{"ads.example.com"},
			RestrictCSS:  []string{"main"},
			Tags:         []string{"a", "area"},
			Attrs:        []string{"href", "data-url"},
			Process: func(raw string) (string, bool) {
				if strings.Contains(raw, "/asset") {
					return raw + "?b=2&a=1", true
				}
				return raw, true
			},
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor returned error: %v", err)
		}
		links, err := extractor.Extract(response)
		if err != nil {
			t.Fatalf("Extract returned error: %v", err)
		}
		want := []string{
			"https://example.com/keep?a=1&b=2",
			"https://blog.example.com/post/1",
			"https://example.com/area",
			"https://example.com/asset?a=1&b=2",
		}
		if !reflect.DeepEqual(links, want) {
			t.Fatalf("links = %#v, want %#v", links, want)
		}
		if !extractor.Matches("https://blog.example.com/post/1") {
			t.Fatal("Matches should accept allowed subdomain URL")
		}
		if extractor.Matches("https://example.com/file.pdf") {
			t.Fatal("Matches should reject ignored extension")
		}

		opaqueExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{})
		if err != nil {
			t.Fatalf("NewLinkExtractor opaque returned error: %v", err)
		}
		opaqueResponse := newTemplateResponse(t, "https://example.com/base/index.html", `<main><a href="http:relative">bad</a><a href="/keep">good</a></main>`, "text/html")
		opaqueLinks, err := opaqueExtractor.Extract(opaqueResponse)
		if err != nil {
			t.Fatalf("Extract opaque HTTP candidate returned error: %v", err)
		}
		if !reflect.DeepEqual(opaqueLinks, []string{"https://example.com/keep"}) {
			t.Fatalf("opaque HTTP candidate links = %#v, want only valid host-backed HTTP URL", opaqueLinks)
		}
		if opaqueExtractor.Matches("http:relative") {
			t.Fatal("Matches should reject opaque HTTP URLs without a host")
		}

		fragmentExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Allow:        []string{`/keep`},
			RestrictCSS:  []string{"main"},
			KeepFragment: true,
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor fragment returned error: %v", err)
		}
		fragmentLinks, err := fragmentExtractor.Extract(response)
		if err != nil {
			t.Fatalf("Extract fragment returned error: %v", err)
		}
		if fragmentLinks[0] != "https://example.com/keep?a=1&b=2#frag" {
			t.Fatalf("fragment link = %q", fragmentLinks[0])
		}

		xpathExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Allow:         []string{`xpath-keep`},
			RestrictXPath: []string{`//section[@class="xpath-only"]`},
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor xpath returned error: %v", err)
		}
		xpathLinks, err := xpathExtractor.Extract(response)
		if err != nil {
			t.Fatalf("Extract xpath returned error: %v", err)
		}
		if !reflect.DeepEqual(xpathLinks, []string{"https://example.com/xpath-keep"}) {
			t.Fatalf("xpath links = %#v", xpathLinks)
		}

		combinedExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Allow:         []string{`scope-`},
			RestrictCSS:   []string{"section.css-scope"},
			RestrictXPath: []string{`//section[@class="xpath-scope"]`},
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor combined returned error: %v", err)
		}
		combinedResponse := newTemplateResponse(t, "https://example.com/base/index.html", `<main><section class="css-scope"><a href="/scope-css">css</a></section><section class="xpath-scope"><a href="/scope-xpath">xpath</a></section></main>`, "text/html")
		combinedLinks, err := combinedExtractor.Extract(combinedResponse)
		if err != nil {
			t.Fatalf("Extract combined returned error: %v", err)
		}
		if !reflect.DeepEqual(combinedLinks, []string{"https://example.com/scope-xpath", "https://example.com/scope-css"}) {
			t.Fatalf("combined restricted links = %#v, want XPath scope before CSS scope", combinedLinks)
		}

		missingScopeExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Allow:       []string{`/fallback`},
			RestrictCSS: []string{"aside.missing"},
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor missing scope returned error: %v", err)
		}
		missingScopeResponse := newTemplateResponse(t, "https://example.com/base/index.html", `<main><a href="/fallback">fallback</a></main>`, "text/html")
		missingScopeLinks, err := missingScopeExtractor.Extract(missingScopeResponse)
		if err != nil {
			t.Fatalf("Extract missing scope returned error: %v", err)
		}
		if !reflect.DeepEqual(missingScopeLinks, []string{"https://example.com/fallback"}) {
			t.Fatalf("missing restricted scope links = %#v, want whole-response fallback", missingScopeLinks)
		}
	})

	t.Run("link extractor canonicalizes hosts without lowercasing URL userinfo", func(t *testing.T) {
		credentialsResponse := newTemplateResponse(t, "https://example.com/index.html", `<a href="https://User:Secret@Example.COM/path">credentials</a>`, "text/html")
		credentialsExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{})
		if err != nil {
			t.Fatalf("NewLinkExtractor credentials returned error: %v", err)
		}
		credentialsLinks, err := credentialsExtractor.Extract(credentialsResponse)
		if err != nil {
			t.Fatalf("Extract credentials returned error: %v", err)
		}
		if !reflect.DeepEqual(credentialsLinks, []string{"https://User:Secret@example.com/path"}) {
			t.Fatalf("credential link = %#v, want userinfo case preserved with normalized host", credentialsLinks)
		}
	})

	t.Run("link extractor resolves candidates against document base href", func(t *testing.T) {
		baseResponse := newTemplateResponse(t, "https://example.com/dir/page.html", `<html><head><base href="https://cdn.example.org/assets/"></head><body><a href="next">next</a></body></html>`, "text/html")
		baseExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{})
		if err != nil {
			t.Fatalf("NewLinkExtractor base returned error: %v", err)
		}
		baseLinks, err := baseExtractor.Extract(baseResponse)
		if err != nil {
			t.Fatalf("Extract base returned error: %v", err)
		}
		if !reflect.DeepEqual(baseLinks, []string{"https://cdn.example.org/assets/next"}) {
			t.Fatalf("base links = %#v, want document base href resolved links", baseLinks)
		}
	})

	t.Run("link extractor process hook can rewrite to relative URL", func(t *testing.T) {
		processResponse := newTemplateResponse(t, "https://example.com/base/index.html", `<a href="/old">old</a>`, "text/html")
		processExtractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{
			Process: func(raw string) (string, bool) {
				if raw == "https://example.com/old" {
					return "../new?b=2&a=1#drop", true
				}
				return raw, true
			},
		})
		if err != nil {
			t.Fatalf("NewLinkExtractor process rewrite returned error: %v", err)
		}
		links, err := processExtractor.Extract(processResponse)
		if err != nil {
			t.Fatalf("Extract process rewrite returned error: %v", err)
		}
		if !reflect.DeepEqual(links, []string{"https://example.com/new?a=1&b=2"}) {
			t.Fatalf("process rewrite links = %#v, want relative rewrite resolved against page URL", links)
		}
	})

	t.Run("crawl spider rules generate requests with callbacks priority and process hooks", func(t *testing.T) {
		extractor, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{Allow: []string{`/post/`}, AllowDomains: []string{"example.com"}, DenyDomains: []string{"ads.example.com"}, RestrictCSS: []string{"main"}})
		if err != nil {
			t.Fatalf("NewLinkExtractor returned error: %v", err)
		}
		callback := func(context.Context, spiders.Response) ([]spiders.Output, error) { return nil, nil }
		crawl := spidertemplates.CrawlSpider{Rules: []spidertemplates.CrawlRule{{
			LinkExtractor: extractor,
			Callback:      callback,
			Priority:      spiders.Priority(7),
			ProcessRequest: func(request spiders.Request, _ spiders.Response) (spiders.Request, error) {
				request.Headers.Set("X-Rule", "post")
				request.Meta = map[string]any{"rule": "post"}
				return request, nil
			},
		}}}
		outputs, err := crawl.Parse(context.Background(), response)
		if err != nil {
			t.Fatalf("CrawlSpider Parse returned error: %v", err)
		}
		if len(outputs) != 1 || outputs[0].Request == nil {
			t.Fatalf("outputs = %#v", outputs)
		}
		request := outputs[0].Request
		if request.URL != "https://blog.example.com/post/1" || request.Priority != 7 || request.Callback == nil {
			t.Fatalf("request = %#v", request)
		}
		if request.Headers.Get("X-Rule") != "post" || request.Meta["rule"] != "post" {
			t.Fatalf("processed request = %#v", request)
		}
	})

	t.Run("sitemap spider follows indexes robots directives alternates and first matching rules", func(t *testing.T) {
		follow, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{Allow: []string{`posts-sitemap`}})
		if err != nil {
			t.Fatalf("NewLinkExtractor follow returned error: %v", err)
		}
		sitemap := &spidertemplates.SitemapSpider{SitemapFollow: follow}
		indexOutputs, err := sitemap.ParseSitemap(context.Background(), newTemplateResponse(t, "https://example.com/sitemap-index.xml", templateFixture(t, "sitemap-index.xml"), "application/xml"))
		if err != nil {
			t.Fatalf("ParseSitemap index returned error: %v", err)
		}
		if got := requestURLs(indexOutputs); !reflect.DeepEqual(got, []string{"https://example.com/posts-sitemap.xml"}) {
			t.Fatalf("index requests = %#v", got)
		}
		if indexOutputs[0].Request.Callback == nil {
			t.Fatal("child sitemap request should recurse through ParseSitemap callback")
		}

		robotsOutputs, err := sitemap.ParseSitemap(context.Background(), newTemplateResponse(t, "https://example.com/robots.txt", templateFixture(t, "robots.txt"), "text/plain"))
		if err != nil {
			t.Fatalf("ParseSitemap robots returned error: %v", err)
		}
		if got := requestURLs(robotsOutputs); !reflect.DeepEqual(got, []string{"https://example.com/sitemap-index.xml", "https://example.com/other-sitemap.xml"}) {
			t.Fatalf("robots sitemap requests = %#v", got)
		}

		posts, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{Allow: []string{`/posts/`}})
		if err != nil {
			t.Fatalf("NewLinkExtractor posts returned error: %v", err)
		}
		products, err := spiders.NewLinkExtractor(spiders.LinkExtractorOptions{Allow: []string{`/products/`}})
		if err != nil {
			t.Fatalf("NewLinkExtractor products returned error: %v", err)
		}
		sitemap = &spidertemplates.SitemapSpider{
			AlternateLinks: true,
			Rules: []spidertemplates.CrawlRule{
				{LinkExtractor: posts, Priority: spiders.Priority(5)},
				{LinkExtractor: products, ProcessRequest: func(request spiders.Request, _ spiders.Response) (spiders.Request, error) {
					request.Meta = map[string]any{"type": "product"}
					return request, nil
				}},
			},
		}
		urlsetOutputs, err := sitemap.ParseSitemap(context.Background(), newTemplateResponse(t, "https://example.com/posts-sitemap.xml", templateFixture(t, "urlset.xml"), "application/xml"))
		if err != nil {
			t.Fatalf("ParseSitemap urlset returned error: %v", err)
		}
		gotURLs := requestURLs(urlsetOutputs)
		wantURLs := []string{"https://example.com/posts/one", "https://example.com/es/posts/one", "https://example.com/products/two"}
		if !reflect.DeepEqual(gotURLs, wantURLs) {
			t.Fatalf("urlset requests = %#v, want %#v", gotURLs, wantURLs)
		}
		if urlsetOutputs[0].Request.Priority != 5 || urlsetOutputs[1].Request.Priority != 5 {
			t.Fatalf("post priorities = %d/%d, want 5", urlsetOutputs[0].Request.Priority, urlsetOutputs[1].Request.Priority)
		}
		if urlsetOutputs[2].Request.Meta["type"] != "product" {
			t.Fatalf("product request meta = %#v", urlsetOutputs[2].Request.Meta)
		}
	})
}

func newTemplateResponse(t *testing.T, rawURL, body, contentType string) spiders.Response {
	t.Helper()
	response, err := goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        rawURL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{contentType}},
		Request: goscrapling.RequestMetadata{
			Method: http.MethodGet,
			URL:    rawURL,
		},
	})
	if err != nil {
		t.Fatalf("new response: %v", err)
	}
	return spiders.Response{Response: response, Request: spiders.Request{URL: rawURL}}
}

func templateFixture(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("..", "testdata", "spiders", "templates", name))
	if err != nil {
		t.Fatalf("read template fixture %s: %v", name, err)
	}
	return string(body)
}

func requestURLs(outputs []spiders.Output) []string {
	urls := make([]string, 0, len(outputs))
	for _, output := range outputs {
		if output.Request != nil {
			urls = append(urls, output.Request.URL)
		}
	}
	return urls
}
