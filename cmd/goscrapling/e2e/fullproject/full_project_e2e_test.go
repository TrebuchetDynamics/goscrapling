package fullproject_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/cmd/goscrapling/internal/clitest"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes"
	"github.com/TrebuchetDynamics/goscrapling/integrations/mcp"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestFullProjectHermeticEndToEnd(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/catalog":
			if got := r.Header.Get("X-Full-Project"); got == "" {
				t.Fatalf("missing X-Full-Project header for %s", r.URL.Path)
			}
			fmt.Fprint(w, `<!doctype html><html><head><title>Trail Catalog</title></head><body><main>
<article class="product" data-sku="trail-kit"><h2>Trail Kit</h2><a class="detail" href="/detail/trail-kit">Detail</a><span class="price">$42</span></article>
<article class="product" data-sku="camp-mug"><h2>Camp Mug</h2><a class="detail" href="/detail/camp-mug">Detail</a><span class="price">$12</span></article>
</main></body></html>`)
		case "/catalog-redesign":
			fmt.Fprint(w, `<!doctype html><html><body><main>
<section class="card" data-sku="trail-kit"><h2>Trail Kit</h2><strong class="cost">$42</strong></section>
</main></body></html>`)
		case "/detail/trail-kit":
			fmt.Fprint(w, `<!doctype html><html><body><main><h1>Trail Kit Detail</h1><p>Ready for camp.</p></main></body></html>`)
		case "/detail/camp-mug":
			fmt.Fprint(w, `<!doctype html><html><body><main><h1>Camp Mug Detail</h1><p>Enamel mug.</p></main></body></html>`)
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

	session, err := fetchers.NewFetcherSession(fetchers.FetcherSessionOptions{
		Client:  server.Client(),
		Headers: http.Header{"X-Full-Project": []string{"session"}},
		Store:   store,
	})
	if err != nil {
		t.Fatalf("NewFetcherSession: %v", err)
	}
	concurrent := fetchers.NewConcurrentFetcher(fetchers.ConcurrentFetcherOptions{Session: session, MaxConcurrency: 2})
	results := concurrent.Fetch(ctx, []fetchers.ConcurrentRequest{
		{URL: server.URL + "/catalog"},
		{URL: server.URL + "/detail/trail-kit"},
	})
	if len(results) != 2 {
		t.Fatalf("concurrent results len = %d, want 2", len(results))
	}
	for i, result := range results {
		if result.Err != nil {
			t.Fatalf("concurrent result %d error: %v", i, result.Err)
		}
	}
	if got := strings.TrimSpace(results[0].Response.CSS(".product h2::text").Text()); got != "Trail Kit\nCamp Mug" {
		t.Fatalf("concurrent catalog titles = %q", got)
	}

	doc, err := goscrapling.Parse(bytes.NewReader(results[0].Response.Body()), goscrapling.ParseOptions{URL: server.URL + "/catalog", Store: store})
	if err != nil {
		t.Fatalf("Parse catalog: %v", err)
	}
	price, err := doc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "full-project-price", AutoSave: true})
	if err != nil {
		t.Fatalf("save adaptive price: %v", err)
	}
	if got := strings.TrimSpace(price.Text()); got != "$42\n$12" {
		t.Fatalf("saved adaptive price text = %q", got)
	}

	reloaded, err := goscrapling.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("reload FileStore: %v", err)
	}
	redesigned, err := session.Get(server.URL+"/catalog-redesign", fetchers.RequestOptions{Store: reloaded})
	if err != nil {
		t.Fatalf("fetch redesigned catalog: %v", err)
	}
	redesignedDoc, err := goscrapling.Parse(bytes.NewReader(redesigned.Body()), goscrapling.ParseOptions{URL: server.URL + "/catalog", Store: reloaded})
	if err != nil {
		t.Fatalf("Parse redesigned catalog: %v", err)
	}
	relocated, err := redesignedDoc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "full-project-price", Adaptive: true, Percentage: 50})
	if err != nil {
		t.Fatalf("relocate adaptive price: %v", err)
	}
	if got := strings.TrimSpace(relocated.Text()); got != "$42" {
		t.Fatalf("relocated adaptive price = %q", got)
	}

	manager := spiders.NewSessionManager()
	if err := manager.Add("static", spiders.NewStaticSessionAdapter(session, spiders.StaticSessionAdapterOptions{}), spiders.SessionOptions{Default: true}); err != nil {
		t.Fatalf("add static spider session: %v", err)
	}
	parseDetail := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		return []spiders.Output{spiders.Item(map[string]any{"title": response.CSS("h1::text").Text(), "source": response.Meta["source"]})}, nil
	}
	parseCatalog := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		follow, err := response.Follow("/detail/trail-kit", spiders.FollowOptions{Callback: parseDetail, Meta: map[string]any{"source": "catalog"}})
		if err != nil {
			return nil, err
		}
		return []spiders.Output{spiders.Next(follow)}, nil
	}
	crawl, err := (spiders.Crawler{Sessions: manager, DefaultCallback: parseCatalog}).Run(ctx, []spiders.Request{{URL: server.URL + "/catalog", Headers: http.Header{"X-Full-Project": []string{"spider"}}}})
	if err != nil {
		t.Fatalf("spider crawl: %v", err)
	}
	if len(crawl.Items) != 1 || crawl.Items[0]["title"] != "Trail Kit Detail" || crawl.Items[0]["source"] != "catalog" {
		t.Fatalf("spider items = %#v", crawl.Items)
	}

	binary := clitest.BuildBinary(t)
	outputPath := filepath.Join(t.TempDir(), "catalog.txt")
	result := clitest.RunBinary(t, binary, "extract", "get", server.URL+"/catalog", outputPath, "--css-selector", ".product h2::text", "-H", "X-Full-Project: cli")
	if result.Err != nil {
		t.Fatalf("CLI extract: %v\nstdout: %s\nstderr: %s", result.Err, result.Stdout, result.Stderr)
	}
	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read CLI output: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != "Trail Kit\nCamp Mug" {
		t.Fatalf("CLI output = %q", got)
	}

	gormesResult, err := (gormes.StaticExtractionAdapter{Fetcher: fetchers.Fetcher{Client: server.Client()}}).Call(ctx, gormes.StaticToolCall{
		Tool:        gormes.ToolWebExtract,
		URLs:        []string{server.URL + "/catalog"},
		CSSSelector: "article.product[data-sku='trail-kit']",
		Opts:        fetchers.RequestOptions{Headers: http.Header{"X-Full-Project": []string{"gormes"}}},
	})
	if err != nil {
		t.Fatalf("Gormes static tool: %v", err)
	}
	if len(gormesResult.Results) != 1 || !strings.Contains(gormesResult.Results[0].Content, "Trail Kit") || strings.Contains(gormesResult.Results[0].Content, "Camp Mug") {
		t.Fatalf("Gormes result = %#v", gormesResult.Results)
	}

	mcpServer := mcp.NewServer(mcp.ServerOptions{
		Static: mcp.StaticClientFunc(func(ctx context.Context, rawURL string, opts fetchers.RequestOptions) (*goscrapling.Response, error) {
			opts.Context = ctx
			opts.Headers = http.Header{"X-Full-Project": []string{"mcp"}}
			return (fetchers.Fetcher{Client: server.Client()}).Get(rawURL, opts)
		}),
		BrowserFactory: func(context.Context, mcp.SessionType, browser.BrowserOptions, int) (mcp.BrowserSession, error) {
			return fullProjectBrowserSession{}, nil
		},
	})
	mcpStatic, err := mcpServer.Get(ctx, mcp.GetRequest{URL: server.URL + "/catalog", CSSSelector: "article.product[data-sku='camp-mug'] h2", ExtractionType: mcp.ExtractionText})
	if err != nil {
		t.Fatalf("MCP get: %v", err)
	}
	if len(mcpStatic.Content) != 1 || mcpStatic.Content[0] != "Camp Mug" {
		t.Fatalf("MCP static content = %#v", mcpStatic.Content)
	}
	mcpRendered, err := mcpServer.Fetch(ctx, mcp.FetchRequest{URL: server.URL + "/app", ExtractionType: mcp.ExtractionText})
	if err != nil {
		t.Fatalf("MCP browser fetch: %v", err)
	}
	if len(mcpRendered.Content) != 1 || mcpRendered.Content[0] != "Rendered Project App" {
		t.Fatalf("MCP rendered content = %#v", mcpRendered.Content)
	}
}

type fullProjectBrowserSession struct{}

func (fullProjectBrowserSession) Fetch(_ context.Context, rawURL string, _ browser.BrowserOptions) (*goscrapling.Response, error) {
	return goscrapling.NewResponse(strings.NewReader(`<!doctype html><html><body><main><h1>Rendered Project App</h1></main></body></html>`), goscrapling.ResponseOptions{
		URL:        rawURL + "#rendered",
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	})
}

func (fullProjectBrowserSession) Screenshot(context.Context, string, browser.BrowserOptions) ([]byte, string, error) {
	return []byte("full-project-shot"), "image/png", nil
}

func (fullProjectBrowserSession) Close(context.Context) error { return nil }

func (fullProjectBrowserSession) Alive() bool { return true }
