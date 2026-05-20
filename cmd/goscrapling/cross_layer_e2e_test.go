package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestGoscraplingCrossLayerLocalEndToEnd(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/catalog":
			if got := r.Header.Get("X-Cross-Layer"); got != "fetcher" && got != "cli" && got != "spider" {
				t.Fatalf("X-Cross-Layer header = %q", got)
			}
			fmt.Fprint(w, `<!doctype html><html><body><main>
<article class="product"><h2>Trail Kit</h2><a class="detail" href="/detail">Detail</a><span class="price">$42</span></article>
</main></body></html>`)
		case "/catalog-redesigned":
			fmt.Fprint(w, `<!doctype html><html><body><main>
<section class="card"><h2>Trail Kit</h2><strong class="cost">$42</strong><a class="detail" href="/detail">Detail</a></section>
</main></body></html>`)
		case "/detail":
			fmt.Fprint(w, `<!doctype html><html><head><meta property="og:title" content="Trail Kit Detail"></head><body><main><h1>Trail Kit Detail</h1><p>Ready for camp.</p></main></body></html>`)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	storePath := filepath.Join(t.TempDir(), "adaptive.json")
	store, err := goscrapling.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}

	fetcher := goscrapling.Fetcher{}
	response, err := fetcher.Get(server.URL+"/catalog", goscrapling.RequestOptions{
		Headers: http.Header{"X-Cross-Layer": []string{"fetcher"}},
		Store:   store,
	})
	if err != nil {
		t.Fatalf("fetch catalog: %v", err)
	}
	if response.StatusCode() != http.StatusOK || response.URL() != server.URL+"/catalog" {
		t.Fatalf("response metadata status=%d url=%q", response.StatusCode(), response.URL())
	}
	if got := strings.TrimSpace(response.CSS(".product h2::text").Text()); got != "Trail Kit" {
		t.Fatalf("response selector title = %q", got)
	}

	initialDoc, err := goscrapling.Parse(bytes.NewReader(response.Body()), goscrapling.ParseOptions{URL: response.URL(), Store: store})
	if err != nil {
		t.Fatalf("parse initial document: %v", err)
	}
	initialPrice, err := initialDoc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "catalog-price", AutoSave: true})
	if err != nil {
		t.Fatalf("save adaptive price: %v", err)
	}
	if got := strings.TrimSpace(initialPrice.Text()); got != "$42" {
		t.Fatalf("initial adaptive price = %q", got)
	}

	reloadedStore, err := goscrapling.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("reload FileStore: %v", err)
	}
	changed, err := fetcher.Get(server.URL+"/catalog-redesigned", goscrapling.RequestOptions{Store: reloadedStore})
	if err != nil {
		t.Fatalf("fetch changed catalog: %v", err)
	}
	changedDoc, err := goscrapling.Parse(bytes.NewReader(changed.Body()), goscrapling.ParseOptions{URL: server.URL + "/catalog", Store: reloadedStore})
	if err != nil {
		t.Fatalf("parse changed document: %v", err)
	}
	relocated, err := changedDoc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "catalog-price", Adaptive: true, Percentage: 50})
	if err != nil {
		t.Fatalf("relocate adaptive price: %v", err)
	}
	if got := strings.TrimSpace(relocated.Text()); got != "$42" {
		t.Fatalf("relocated adaptive price = %q", got)
	}

	sessions := spiders.NewSessionManager()
	if err := sessions.Add("static", crossLayerHTTPSession{fetcher: fetcher}, spiders.SessionOptions{Default: true}); err != nil {
		t.Fatalf("add spider session: %v", err)
	}
	parseDetail := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		return []spiders.Output{spiders.Item(map[string]any{
			"title": response.CSS("h1::text").Text(),
			"from":  response.Meta["from"],
		})}, nil
	}
	parseCatalog := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		follow, err := response.Follow("/detail", spiders.FollowOptions{
			Callback: parseDetail,
			Meta:     map[string]any{"from": "catalog"},
			Headers:  http.Header{"X-Cross-Layer": []string{"spider"}},
		})
		if err != nil {
			return nil, err
		}
		return []spiders.Output{spiders.Next(follow)}, nil
	}
	crawl, err := (spiders.Crawler{Sessions: sessions, DefaultCallback: parseCatalog}).Run(ctx, []spiders.Request{{
		URL:     server.URL + "/catalog",
		Headers: http.Header{"X-Cross-Layer": []string{"spider"}},
	}})
	if err != nil {
		t.Fatalf("spider run: %v", err)
	}
	wantItems := []map[string]any{{"title": "Trail Kit Detail", "from": "catalog"}}
	if !reflect.DeepEqual(crawl.Items, wantItems) {
		t.Fatalf("spider items = %#v, want %#v", crawl.Items, wantItems)
	}

	binary := buildGoscraplingBinary(t)
	outputPath := filepath.Join(t.TempDir(), "cli-title.txt")
	result := runGoscraplingBinary(t, binary,
		"extract", "get", server.URL+"/catalog", outputPath,
		"--css-selector", ".product h2::text",
		"-H", "X-Cross-Layer: cli",
	)
	if result.err != nil {
		t.Fatalf("cli extract: %v\nstdout: %s\nstderr: %s", result.err, result.stdout, result.stderr)
	}
	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read CLI output: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != "Trail Kit" {
		t.Fatalf("CLI output = %q", got)
	}

	adapter := gormes.BrowserExtractionAdapter{Engine: crossLayerBrowserEngine{body: []byte(`<!doctype html><html><body><main><h1>Rendered Trail Kit</h1><p>Read the <a href="/detail">detail</a>.</p></main></body></html>`)}}
	toolResult, err := adapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolRenderedMarkdown, URL: server.URL + "/dynamic"})
	if err != nil {
		t.Fatalf("gormes browser tool: %v", err)
	}
	if toolResult.URL != server.URL+"/dynamic-rendered" || toolResult.StatusCode != http.StatusOK {
		t.Fatalf("gormes metadata = %#v", toolResult)
	}
	if got := toolResult.Markdown; got != "# Rendered Trail Kit\n\nRead the [detail](/detail)." {
		t.Fatalf("gormes markdown = %q", got)
	}
}

type crossLayerHTTPSession struct {
	fetcher goscrapling.Fetcher
}

func (s crossLayerHTTPSession) Fetch(_ context.Context, request spiders.Request) (*goscrapling.Response, error) {
	return s.fetcher.Get(request.URL, goscrapling.RequestOptions{Headers: request.Headers})
}

type crossLayerBrowserEngine struct {
	body []byte
}

func (e crossLayerBrowserEngine) Fetch(_ context.Context, request browser.BrowserRequest) (browser.BrowserResult, error) {
	return browser.BrowserResult{
		URL:        request.URL + "-rendered",
		StatusCode: http.StatusOK,
		Reason:     http.StatusText(http.StatusOK),
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       e.body,
	}, nil
}
