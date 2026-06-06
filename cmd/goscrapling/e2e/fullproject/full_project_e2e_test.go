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

type fullProjectRenderedEngine struct{ body []byte }

func (e fullProjectRenderedEngine) Fetch(_ context.Context, request browser.BrowserRequest) (browser.BrowserResult, error) {
	return browser.BrowserResult{
		URL:        request.URL + "#rendered",
		StatusCode: http.StatusOK,
		Reason:     http.StatusText(http.StatusOK),
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Body:       append([]byte(nil), e.body...),
	}, nil
}

type fullProjectStatefulBrowserSession struct {
	body  []byte
	alive bool
}

func (s *fullProjectStatefulBrowserSession) Fetch(_ context.Context, rawURL string, _ browser.BrowserOptions) (*goscrapling.Response, error) {
	return goscrapling.NewResponse(bytes.NewReader(s.body), goscrapling.ResponseOptions{
		URL:        rawURL + "#mcp",
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	})
}

func (s *fullProjectStatefulBrowserSession) Screenshot(_ context.Context, rawURL string, _ browser.BrowserOptions) ([]byte, string, error) {
	return []byte("mcp-shot:" + rawURL), rawURL + "#shot", nil
}

func (s *fullProjectStatefulBrowserSession) Close(context.Context) error {
	s.alive = false
	return nil
}

func (s *fullProjectStatefulBrowserSession) Alive() bool { return s.alive }

func fullProjectHasSemanticRole(nodes []browser.SemanticNode, role, name string) bool {
	for _, node := range nodes {
		if node.Role == role && node.Name == name {
			return true
		}
	}
	return false
}

func (fullProjectBrowserSession) Fetch(_ context.Context, rawURL string, _ browser.BrowserOptions) (*goscrapling.Response, error) {
	return goscrapling.NewResponse(strings.NewReader(`<!doctype html><html><body><main><h1>Rendered Project App</h1></main></body></html>`), goscrapling.ResponseOptions{
		URL:        rawURL + "#rendered",
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	})
}

func TestFullProjectRenderedBrowserToolsEndToEnd(t *testing.T) {
	ctx := context.Background()
	renderedHTML := []byte(`<!doctype html>
<html><head>
<meta property="og:title" content="Rendered Gear App">
<script type="application/ld+json">{"@type":"Product","name":"Trail Kit"}</script>
</head><body><main>
<h1>Rendered Gear App</h1>
<p>Read the <a href="/gear/trail-kit">Trail Kit</a> guide.</p>
<form action="/search"><label for="q">Search gear</label><input id="q" name="q" value="trail"></form>
<button>Add to pack</button>
<label for="stock">In stock</label><input id="stock" type="checkbox" checked>
</main></body></html>`)

	browserAdapter := gormes.BrowserExtractionAdapter{Engine: fullProjectRenderedEngine{body: renderedHTML}}
	markdown, err := browserAdapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolRenderedMarkdown, URL: "https://fixture.local/app"})
	if err != nil {
		t.Fatalf("Gormes rendered markdown: %v", err)
	}
	if !strings.Contains(markdown.Markdown, "# Rendered Gear App") || !strings.Contains(markdown.Markdown, "[Trail Kit](/gear/trail-kit)") {
		t.Fatalf("rendered markdown = %q", markdown.Markdown)
	}
	links, err := browserAdapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolLinks, URL: "https://fixture.local/app"})
	if err != nil {
		t.Fatalf("Gormes links: %v", err)
	}
	if len(links.Links) != 1 || links.Links[0].Text != "Trail Kit" || links.Links[0].Href != "/gear/trail-kit" {
		t.Fatalf("links = %#v", links.Links)
	}
	structured, err := browserAdapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolStructuredData, URL: "https://fixture.local/app"})
	if err != nil {
		t.Fatalf("Gormes structured data: %v", err)
	}
	if structured.StructuredData.OpenGraph["title"] != "Rendered Gear App" || len(structured.StructuredData.JSONLD) != 1 || fmt.Sprint(structured.StructuredData.JSONLD[0]["@type"]) != "Product" {
		t.Fatalf("structured data = %#v", structured.StructuredData)
	}
	semantic, err := browserAdapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolSemanticTree, URL: "https://fixture.local/app"})
	if err != nil {
		t.Fatalf("Gormes semantic tree: %v", err)
	}
	if !fullProjectHasSemanticRole(semantic.SemanticTree, "heading", "Rendered Gear App") || !fullProjectHasSemanticRole(semantic.SemanticTree, "textbox", "Search gear") {
		t.Fatalf("semantic tree = %#v", semantic.SemanticTree)
	}
	interactive, err := browserAdapter.Call(ctx, gormes.ToolCall{Tool: gormes.ToolInteractiveElements, URL: "https://fixture.local/app"})
	if err != nil {
		t.Fatalf("Gormes interactive elements: %v", err)
	}
	if !fullProjectHasSemanticRole(interactive.InteractiveElements, "link", "Trail Kit") || !fullProjectHasSemanticRole(interactive.InteractiveElements, "button", "Add to pack") || !fullProjectHasSemanticRole(interactive.InteractiveElements, "checkbox", "In stock") {
		t.Fatalf("interactive elements = %#v", interactive.InteractiveElements)
	}

	session := &fullProjectStatefulBrowserSession{body: renderedHTML, alive: true}
	mcpServer := mcp.NewServer(mcp.ServerOptions{
		BrowserFactory: func(context.Context, mcp.SessionType, browser.BrowserOptions, int) (mcp.BrowserSession, error) {
			return session, nil
		},
	})
	created, err := mcpServer.OpenSession(ctx, mcp.OpenSessionRequest{SessionType: mcp.SessionDynamic, SessionID: "rendered-e2e", Headless: true, MaxPages: 1})
	if err != nil {
		t.Fatalf("MCP open session: %v", err)
	}
	if created.SessionID != "rendered-e2e" || created.SessionType != mcp.SessionDynamic || !created.IsAlive {
		t.Fatalf("created session = %#v", created)
	}
	mcpRendered, err := mcpServer.Fetch(ctx, mcp.FetchRequest{URL: "https://fixture.local/app", SessionID: "rendered-e2e", CSSSelector: "h1", ExtractionType: mcp.ExtractionText})
	if err != nil {
		t.Fatalf("MCP session fetch: %v", err)
	}
	if mcpRendered.URL != "https://fixture.local/app#mcp" || len(mcpRendered.Content) != 1 || mcpRendered.Content[0] != "Rendered Gear App" {
		t.Fatalf("MCP rendered fetch = %#v", mcpRendered)
	}
	screenshot, err := mcpServer.Screenshot(ctx, mcp.ScreenshotRequest{URL: "https://fixture.local/app", SessionID: "rendered-e2e", ImageType: "png", FullPage: true})
	if err != nil {
		t.Fatalf("MCP screenshot: %v", err)
	}
	if len(screenshot) != 2 || screenshot[0].Type != mcp.ContentImage || screenshot[0].MimeType != "image/png" || string(screenshot[0].Data) != "mcp-shot:https://fixture.local/app" || screenshot[1].Text != "https://fixture.local/app#shot" {
		t.Fatalf("MCP screenshot blocks = %#v", screenshot)
	}
	closed, err := mcpServer.CloseSession(ctx, mcp.CloseSessionRequest{SessionID: "rendered-e2e"})
	if err != nil {
		t.Fatalf("MCP close session: %v", err)
	}
	if closed.SessionID != "rendered-e2e" || session.Alive() {
		t.Fatalf("closed session = %#v alive=%v", closed, session.Alive())
	}
	infos, err := mcpServer.ListSessions(ctx)
	if err != nil {
		t.Fatalf("MCP list sessions after close: %v", err)
	}
	if len(infos) != 0 {
		t.Fatalf("sessions after close = %#v", infos)
	}
}

func TestFullProjectGormesProviderRecipeContainerEndToEnd(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "trail" {
			t.Fatalf("query = %q, want trail", got)
		}
		if got := r.URL.Query().Get("p"); got != "11" {
			t.Fatalf("page = %q, want 11", got)
		}
		if got := r.Header.Get("X-Gormes-E2E"); got != "trail" {
			t.Fatalf("recipe header = %q, want trail", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"results": [{
				"url": "/gear/trail-kit",
				"title": "<b>Trail Kit</b>",
				"summary": "<p>Ready for <em>camp</em></p>"
			}],
			"suggestions": ["trail kit", "camp gear", "trail kit"]
		}`)
	}))
	t.Cleanup(server.Close)

	adapter := gormes.StaticExtractionAdapter{
		Fetcher: fetchers.Fetcher{Client: server.Client()},
		Recipes: map[string]gormes.ExtractionRecipe{
			"gear_search": {
				URLTemplate: server.URL + "/search?q={query}&p={pageno}",
				Headers:     http.Header{"X-Gormes-E2E": []string{"{query}"}},
				PageSize:    10,
				Fields: []gormes.ExtractionField{
					{Name: "url", Type: gormes.SelectorJSON, Selector: "results/url", Prefix: server.URL},
					{Name: "title", Type: gormes.SelectorJSON, Selector: "results/title", HTMLToText: true},
					{Name: "summary", Type: gormes.SelectorJSON, Selector: "results/summary", HTMLToText: true},
				},
				Suggestions: []gormes.ExtractionField{{Type: gormes.SelectorJSON, Selector: "suggestions", Multiple: true}},
			},
		},
		Providers: []gormes.StaticProvider{
			{Name: "gear", Shortcut: "g", Categories: []string{"shopping", "it"}, Enabled: true, Weight: 4, Recipe: "gear_search"},
			{Name: "mirror", Shortcut: "m", Categories: []string{"shopping"}, Enabled: true, Weight: 2, Recipe: "gear_search"},
		},
	}

	providers := adapter.ProvidersByCategory("shopping")
	if len(providers) != 2 || providers[0].Name != "gear" || providers[1].Name != "mirror" {
		t.Fatalf("providers by category = %#v", providers)
	}
	if provider, err := adapter.ResolveProvider("g"); err != nil || provider.Name != "gear" {
		t.Fatalf("ResolveProvider(g) = %#v, %v", provider, err)
	}

	primary, err := adapter.Call(ctx, gormes.StaticToolCall{
		Tool:         gormes.ToolWebExtract,
		Provider:     "gear",
		RecipeParams: map[string]string{"query": "trail"},
		Pageno:       2,
	})
	if err != nil {
		t.Fatalf("primary provider call: %v", err)
	}
	mirror, err := adapter.Call(ctx, gormes.StaticToolCall{
		Tool:         gormes.ToolWebExtract,
		Provider:     "mirror",
		RecipeParams: map[string]string{"query": "trail"},
		Pageno:       2,
	})
	if err != nil {
		t.Fatalf("mirror provider call: %v", err)
	}

	container := gormes.NewResultContainer()
	container.Add(primary.Results[0])
	container.Add(mirror.Results[0])
	container.AddUnresponsiveProvider("offline", "timeout")
	snapshot := container.Snapshot()
	if len(snapshot.Results) != 1 {
		t.Fatalf("merged results len = %d, want one", len(snapshot.Results))
	}
	merged := snapshot.Results[0]
	if got := merged.Fields["title"].Value; got != "Trail Kit" {
		t.Fatalf("title field = %q", got)
	}
	if got := merged.Fields["url"].Value; got != server.URL+"/gear/trail-kit" {
		t.Fatalf("url field = %q", got)
	}
	if got := merged.Fields["summary"].Value; got != "Ready for camp" {
		t.Fatalf("summary field = %q", got)
	}
	if got := strings.Join(merged.ProviderNames(), ","); got != "gear,mirror" {
		t.Fatalf("merged providers = %q", got)
	}
	if got := merged.Score; got != 5 {
		t.Fatalf("merged score = %v, want 5", got)
	}
	if got := strings.Join(snapshot.Suggestions, ","); got != "trail kit,camp gear" {
		t.Fatalf("snapshot suggestions = %q", got)
	}
	if len(snapshot.UnresponsiveProviders) != 1 || snapshot.UnresponsiveProviders[0].Provider != "offline" || snapshot.UnresponsiveProviders[0].Error != "timeout" {
		t.Fatalf("unresponsive providers = %#v", snapshot.UnresponsiveProviders)
	}
}

func (fullProjectBrowserSession) Screenshot(context.Context, string, browser.BrowserOptions) ([]byte, string, error) {
	return []byte("full-project-shot"), "image/png", nil
}

func (fullProjectBrowserSession) Close(context.Context) error { return nil }

func (fullProjectBrowserSession) Alive() bool { return true }
