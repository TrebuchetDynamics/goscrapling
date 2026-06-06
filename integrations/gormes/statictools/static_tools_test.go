package statictools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling/fetchers"
)

func TestGormesIntegration(t *testing.T) {
	t.Run("static web_extract returns selector content and structured evidence", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/catalog" {
				t.Fatalf("request path = %q, want /catalog", r.URL.Path)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Catalog</title></head>
<body>
  <nav>Navigation should stay out</nav>
  <article class="product">
    <h2>Alpha Tool</h2>
    <p>Useful static extraction target.</p>
  </article>
  <article class="ad">
    <p>Ad copy should not be returned.</p>
  </article>
</body>
</html>`))
		}))
		defer server.Close()

		adapter := StaticExtractionAdapter{Fetcher: fetchers.Fetcher{Client: server.Client()}}
		response, err := adapter.Call(context.Background(), StaticToolCall{
			Tool:        ToolWebExtract,
			URLs:        []string{server.URL + "/catalog"},
			CSSSelector: "article.product",
		})
		if err != nil {
			t.Fatalf("Call returned error: %v", err)
		}
		if len(response.Results) != 1 {
			t.Fatalf("results len = %d, want one", len(response.Results))
		}

		result := response.Results[0]
		if result.URL != server.URL+"/catalog" || result.Title != "Catalog" {
			t.Fatalf("result metadata = %#v, want final URL and title", result)
		}
		for _, want := range []string{"Alpha Tool", "Useful static extraction target."} {
			if !strings.Contains(result.Content, want) {
				t.Fatalf("selected content missing %q:\n%s", want, result.Content)
			}
		}
		for _, forbidden := range []string{"Navigation should stay out", "Ad copy should not be returned"} {
			if strings.Contains(result.Content, forbidden) {
				t.Fatalf("selected content leaked %q:\n%s", forbidden, result.Content)
			}
		}
		if result.Extraction == nil {
			t.Fatal("expected extraction evidence")
		}
		if got := *result.Extraction; got.Engine != "goscrapling" || got.Mode != "static" || got.StatusCode != http.StatusOK || got.ContentType != "text/html" || got.CSSSelector != "article.product" || got.FinalURL != result.URL {
			t.Fatalf("extraction evidence = %#v, want goscrapling static selector evidence", got)
		}

		payload, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		for _, want := range []string{"\"results\"", "\"extraction\"", "\"css_selector\"", "\"content_type\""} {
			if !strings.Contains(string(payload), want) {
				t.Fatalf("JSON payload %s missing %s", payload, want)
			}
		}
	})
}

func TestGormesDeclarativeExtractionRecipes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/catalog/alpha":
			if got := r.URL.Query().Get("q"); got != "tools" {
				t.Fatalf("html recipe query = %q, want tools", got)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Catalog Alpha</title></head>
<body>
  <article class="product">
    <h2>Alpha Tool</h2>
    <p class="summary">Useful static extraction target.</p>
  </article>
</body>
</html>`))
		case "/api/product":
			if got := r.URL.Query().Get("q"); got != "alpha" {
				t.Fatalf("json recipe query = %q, want alpha", got)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"product":{"name":"Alpha Tool","price":"19.99","tags":["go","scraping"]}}`))
		default:
			t.Fatalf("unexpected recipe request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := StaticExtractionAdapter{
		Fetcher: fetchers.Fetcher{Client: server.Client()},
		Recipes: map[string]ExtractionRecipe{
			"catalog_html": {
				URLTemplate: server.URL + "/catalog/{slug}",
				Params:      url.Values{"q": []string{"{query}"}},
				Fields: []ExtractionField{
					{Name: "heading", Type: SelectorCSS, Selector: "article.product h2::text"},
					{Name: "summary", Type: SelectorXPath, Selector: `//article[@class="product"]/p/text()`},
				},
			},
			"catalog_json": {
				URLTemplate: server.URL + "/api/product",
				Params:      url.Values{"q": []string{"{slug}"}},
				Fields: []ExtractionField{
					{Name: "name", Type: SelectorJSON, Selector: "product/name"},
					{Name: "price", Type: SelectorJSON, Selector: "product/price"},
					{Name: "tags", Type: SelectorJSON, Selector: "product/tags", Multiple: true},
				},
			},
			"broken_html": {
				URLTemplate: server.URL + "/catalog/{slug}",
				Params:      url.Values{"q": []string{"{query}"}},
				Fields: []ExtractionField{
					{Name: "missing", Type: SelectorCSS, Selector: "article.missing::text"},
					{Name: "invalid", Type: SelectorXPath, Selector: `//*[`, Required: true},
				},
			},
		},
	}

	htmlResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Recipe:       "catalog_html",
		RecipeParams: map[string]string{"slug": "alpha", "query": "tools"},
	})
	if err != nil {
		t.Fatalf("html recipe call returned error: %v", err)
	}
	if len(htmlResponse.Results) != 1 {
		t.Fatalf("html recipe results = %d, want one", len(htmlResponse.Results))
	}
	htmlResult := htmlResponse.Results[0]
	if htmlResult.URL != server.URL+"/catalog/alpha?q=tools" || htmlResult.Extraction == nil || htmlResult.Extraction.Recipe != "catalog_html" {
		t.Fatalf("html recipe metadata = %#v", htmlResult)
	}
	if got := htmlResult.Fields["heading"].Value; got != "Alpha Tool" {
		t.Fatalf("heading field = %q, want Alpha Tool", got)
	}
	if got := htmlResult.Fields["summary"].Value; got != "Useful static extraction target." {
		t.Fatalf("summary field = %q", got)
	}

	jsonResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Recipe:       "catalog_json",
		RecipeParams: map[string]string{"slug": "alpha"},
	})
	if err != nil {
		t.Fatalf("json recipe call returned error: %v", err)
	}
	jsonResult := jsonResponse.Results[0]
	if jsonResult.Extraction == nil || jsonResult.Extraction.ContentType != "application/json" {
		t.Fatalf("json extraction evidence = %#v", jsonResult.Extraction)
	}
	if got := jsonResult.Fields["name"].Value; got != "Alpha Tool" {
		t.Fatalf("json name field = %q", got)
	}
	if got := jsonResult.Fields["price"].Value; got != "19.99" {
		t.Fatalf("json price field = %q", got)
	}
	if got := strings.Join(jsonResult.Fields["tags"].Values, ","); got != "go,scraping" {
		t.Fatalf("json tags field = %q", got)
	}

	brokenResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Recipe:       "broken_html",
		RecipeParams: map[string]string{"slug": "alpha", "query": "tools"},
	})
	if err != nil {
		t.Fatalf("broken recipe call returned error: %v", err)
	}
	brokenFields := brokenResponse.Results[0].Fields
	if got := brokenFields["missing"].Error; !strings.Contains(got, "matched no values") {
		t.Fatalf("missing selector error = %q", got)
	}
	if got := brokenFields["invalid"].Error; !strings.Contains(got, "invalid xpath") {
		t.Fatalf("invalid selector error = %q", got)
	}
}

func TestGormesRecipeRequestControls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			if r.Method != http.MethodGet {
				t.Fatalf("search method = %s, want GET", r.Method)
			}
			checks := map[string]string{
				"q":    "tools",
				"p":    "11",
				"lang": "en",
				"safe": "strict",
				"age":  "7d",
			}
			for key, want := range checks {
				if got := r.URL.Query().Get(key); got != want {
					t.Fatalf("search query %s = %q, want %q", key, got, want)
				}
			}
			if got := r.Header.Get("X-Recipe-Query"); got != "tools" {
				t.Fatalf("search header = %q, want tools", got)
			}
			if cookie, err := r.Cookie("mode"); err != nil || cookie.Value != "static" {
				t.Fatalf("search cookie = %#v, %v; want mode=static", cookie, err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"results": [{
					"url": "/alpha",
					"title": "<b>Alpha Tool</b>",
					"thumbnail": "/thumb.png",
					"summary": "<p>Useful <em>HTML</em> summary.</p>"
				}],
				"suggestions": ["alpha tools", "go scraping"]
			}`))
		case "/submit":
			if r.Method != http.MethodPost {
				t.Fatalf("submit method = %s, want POST", r.Method)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read submit body: %v", err)
			}
			if got := string(body); got != "search=tools&page=1&lang=en" {
				t.Fatalf("submit body = %q", got)
			}
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Fatalf("submit content type = %q", got)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><head><title>Posted</title></head><body><article><h2>Posted Tool</h2></article></body></html>`))
		case "/empty":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`not found`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	adapter := StaticExtractionAdapter{
		Fetcher: fetchers.Fetcher{Client: server.Client()},
		Recipes: map[string]ExtractionRecipe{
			"search_controls": {
				URLTemplate:       server.URL + "/search?q={query}&p={pageno}&lang={lang}{safe_search}{time_range}",
				Method:            http.MethodGet,
				Headers:           http.Header{"X-Recipe-Query": []string{"{query}"}},
				Cookies:           map[string]string{"mode": "static"},
				PageSize:          10,
				FirstPageNum:      1,
				LangAll:           "en",
				SafeSearchMap:     map[int]string{2: "&safe=strict"},
				TimeRangeTemplate: "&age={time_range_val}",
				TimeRangeMap:      map[string]string{"week": "7d"},
				Fields: []ExtractionField{
					{Name: "url", Type: SelectorJSON, Selector: "results/url", Prefix: server.URL},
					{Name: "title", Type: SelectorJSON, Selector: "results/title", HTMLToText: true},
					{Name: "thumbnail", Type: SelectorJSON, Selector: "results/thumbnail", Prefix: server.URL},
					{Name: "summary", Type: SelectorJSON, Selector: "results/summary", HTMLToText: true},
				},
				Suggestions: []ExtractionField{{Type: SelectorJSON, Selector: "suggestions", Multiple: true}},
			},
			"post_controls": {
				URLTemplate:  server.URL + "/submit",
				Method:       http.MethodPost,
				Headers:      http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}},
				BodyTemplate: "search={query}&page={pageno}&lang={lang}",
				LangAll:      "en",
				Fields: []ExtractionField{
					{Name: "heading", Type: SelectorCSS, Selector: "article h2::text"},
				},
			},
			"empty_status": {
				URLTemplate:         server.URL + "/empty",
				NoResultStatusCodes: []int{http.StatusNotFound},
				Fields: []ExtractionField{
					{Name: "body", Type: SelectorCSS, Selector: "body::text"},
				},
			},
		},
	}

	searchResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Recipe:       "search_controls",
		RecipeParams: map[string]string{"query": "tools"},
		Pageno:       2,
		Language:     "en-US",
		SafeSearch:   2,
		TimeRange:    "week",
	})
	if err != nil {
		t.Fatalf("search recipe call returned error: %v", err)
	}
	searchResult := searchResponse.Results[0]
	if got := searchResult.Fields["url"].Value; got != server.URL+"/alpha" {
		t.Fatalf("url field = %q", got)
	}
	if got := searchResult.Fields["title"].Value; got != "Alpha Tool" {
		t.Fatalf("title field = %q", got)
	}
	if got := searchResult.Fields["thumbnail"].Value; got != server.URL+"/thumb.png" {
		t.Fatalf("thumbnail field = %q", got)
	}
	if got := searchResult.Fields["summary"].Value; got != "Useful HTML summary." {
		t.Fatalf("summary field = %q", got)
	}
	if got := strings.Join(searchResult.Suggestions, ","); got != "alpha tools,go scraping" {
		t.Fatalf("suggestions = %q", got)
	}

	postResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Recipe:       "post_controls",
		RecipeParams: map[string]string{"query": "tools"},
		Language:     "all",
	})
	if err != nil {
		t.Fatalf("post recipe call returned error: %v", err)
	}
	if got := postResponse.Results[0].Fields["heading"].Value; got != "Posted Tool" {
		t.Fatalf("posted heading = %q", got)
	}

	emptyResponse, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:   ToolWebExtract,
		Recipe: "empty_status",
	})
	if err != nil {
		t.Fatalf("empty recipe call returned error: %v", err)
	}
	emptyResult := emptyResponse.Results[0]
	if emptyResult.Error != "" {
		t.Fatalf("empty status error = %q, want no error", emptyResult.Error)
	}
	if emptyResult.Extraction == nil || !emptyResult.Extraction.NoResult {
		t.Fatalf("empty extraction = %#v, want no-result evidence", emptyResult.Extraction)
	}
	if len(emptyResult.Fields) != 0 {
		t.Fatalf("empty fields = %#v, want none", emptyResult.Fields)
	}
}

func TestGormesResultContainerMergeScoring(t *testing.T) {
	container := NewResultContainer()
	container.Add(WebExtractResult{
		URL:     "https://example.test/tool",
		Title:   "Alpha Tool",
		Content: "First provider content.",
		Fields: map[string]ExtractedField{
			"heading": {Type: SelectorCSS, Selector: "h2::text", Value: "Alpha Tool"},
		},
		Suggestions: []string{"alpha", "tools"},
		Extraction:  &WebExtraction{Provider: "docs", ProviderWeight: 2, FinalURL: "https://example.test/tool"},
	})
	container.Add(WebExtractResult{
		URL:     "https://example.test/other",
		Title:   "Other Tool",
		Content: "Other provider content.",
		Fields: map[string]ExtractedField{
			"heading": {Type: SelectorCSS, Selector: "h2::text", Value: "Other Tool"},
		},
		Extraction: &WebExtraction{Provider: "archive", ProviderWeight: 1.5, FinalURL: "https://example.test/other"},
	})
	container.Add(WebExtractResult{
		URL:     "https://example.test/tool",
		Title:   "Alpha Tool Mirror",
		Content: "Duplicate provider content should not create another row.",
		Fields: map[string]ExtractedField{
			"summary": {Type: SelectorCSS, Selector: "p::text", Value: "Duplicate summary"},
		},
		Suggestions: []string{"tools", "scraping"},
		Extraction:  &WebExtraction{Provider: "search", ProviderWeight: 3, FinalURL: "https://example.test/tool"},
	})
	container.AddUnresponsiveProvider("broken", "timeout")

	output := container.Snapshot()
	if len(output.Results) != 2 {
		t.Fatalf("merged results = %d, want two", len(output.Results))
	}
	first := output.Results[0]
	if first.URL != "https://example.test/tool" || first.Title != "Alpha Tool" || first.Content != "First provider content." {
		t.Fatalf("first result = %#v, want merged alpha row", first)
	}
	if first.Score != 3 {
		t.Fatalf("first score = %v, want 3", first.Score)
	}
	if got := strings.Join(first.ProviderNames(), ","); got != "docs,search" {
		t.Fatalf("first providers = %q", got)
	}
	if got := first.Providers[0].Position; got != 1 {
		t.Fatalf("first provider position = %d, want 1", got)
	}
	if got := first.Providers[1].Position; got != 3 {
		t.Fatalf("duplicate provider position = %d, want 3", got)
	}
	if got := first.Providers[1].Fields["summary"].Value; got != "Duplicate summary" {
		t.Fatalf("duplicate provider field evidence = %q", got)
	}
	if got := strings.Join(first.Suggestions, ","); got != "alpha,tools,scraping" {
		t.Fatalf("merged suggestions = %q", got)
	}
	if got := output.Results[1].URL; got != "https://example.test/other" {
		t.Fatalf("second URL = %q, want other result ordered after lower score", got)
	}
	if got := strings.Join(output.Suggestions, ","); got != "alpha,tools,scraping" {
		t.Fatalf("container suggestions = %q", got)
	}
	if len(output.UnresponsiveProviders) != 1 || output.UnresponsiveProviders[0].Provider != "broken" || output.UnresponsiveProviders[0].Error != "timeout" {
		t.Fatalf("unresponsive diagnostics = %#v", output.UnresponsiveProviders)
	}
}

func TestGormesStaticProviderRegistry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/provider/search" {
			t.Fatalf("provider path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "tools" {
			t.Fatalf("provider query = %q, want tools", got)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html>
<head><title>Provider Search</title></head>
<body><article><h2>Provider Tool</h2></article></body>
</html>`))
	}))
	defer server.Close()

	adapter := StaticExtractionAdapter{
		Fetcher: fetchers.Fetcher{Client: server.Client()},
		Recipes: map[string]ExtractionRecipe{
			"docs_recipe": {
				URLTemplate: server.URL + "/provider/search?q={query}",
				Fields: []ExtractionField{
					{Name: "heading", Type: SelectorCSS, Selector: "article h2::text"},
				},
			},
		},
		Providers: []StaticProvider{
			{Name: "docs", Shortcut: "d", Categories: []string{"general", "it"}, Enabled: true, Timeout: 25 * time.Millisecond, Weight: 2.5, Recipe: "docs_recipe"},
			{Name: "disabled", Shortcut: "off", Categories: []string{"general"}, Enabled: false, Recipe: "docs_recipe"},
			{Name: "ambiguous one", Shortcut: "dup", Categories: []string{"general"}, Enabled: true, Recipe: "docs_recipe"},
			{Name: "ambiguous two", Shortcut: "dup", Categories: []string{"general"}, Enabled: true, Recipe: "docs_recipe"},
		},
	}

	providers := adapter.ProvidersByCategory("it")
	if len(providers) != 1 || providers[0].Name != "docs" || providers[0].Shortcut != "d" || providers[0].Weight != 2.5 || providers[0].Timeout != 25*time.Millisecond {
		t.Fatalf("providers by category = %#v, want enabled docs provider", providers)
	}

	provider, err := adapter.ResolveProvider("d")
	if err != nil {
		t.Fatalf("resolve shortcut: %v", err)
	}
	if provider.Name != "docs" {
		t.Fatalf("resolved provider = %#v, want docs", provider)
	}

	if _, err := adapter.ResolveProvider("disabled"); err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("disabled provider error = %v, want disabled", err)
	}
	if _, err := adapter.ResolveProvider("dup"); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous provider error = %v, want ambiguous", err)
	}

	response, err := adapter.Call(context.Background(), StaticToolCall{
		Tool:         ToolWebExtract,
		Provider:     "docs",
		RecipeParams: map[string]string{"query": "tools"},
	})
	if err != nil {
		t.Fatalf("provider call returned error: %v", err)
	}
	if len(response.Results) != 1 {
		t.Fatalf("provider results = %d, want one", len(response.Results))
	}
	result := response.Results[0]
	if got := result.Fields["heading"].Value; got != "Provider Tool" {
		t.Fatalf("provider heading = %q, want Provider Tool", got)
	}
	if result.Extraction == nil || result.Extraction.Provider != "docs" || result.Extraction.ProviderShortcut != "d" || result.Extraction.ProviderWeight != 2.5 || result.Extraction.Recipe != "docs_recipe" {
		t.Fatalf("provider extraction = %#v", result.Extraction)
	}
}
