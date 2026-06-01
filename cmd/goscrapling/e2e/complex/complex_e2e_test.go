package complex_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/cmd/goscrapling/internal/clitest"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/TrebuchetDynamics/goscrapling/spiders"
)

func TestGoscraplingComplexStatefulEndToEnd(t *testing.T) {
	ctx := context.Background()
	var orderBodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		switch r.URL.Path {
		case "/login":
			http.SetCookie(w, &http.Cookie{Name: "sid", Value: "complex-session", Path: "/"})
			fmt.Fprint(w, `<html><body><p class="status">logged in</p></body></html>`)
		case "/catalog":
			if got := r.URL.Query().Get("segment"); got != "gear" {
				t.Fatalf("catalog segment = %q, want gear", got)
			}
			cookieValue := "missing"
			if cookie, err := r.Cookie("sid"); err == nil {
				cookieValue = cookie.Value
			}
			fmt.Fprintf(w, `<!doctype html><html><head><title>Complex Catalog</title></head><body><main data-cookie="%s">
<article class="product" data-sku="trail-kit"><h2>Trail Kit</h2><a class="detail" href="/detail/trail-kit">Detail</a><span class="price">$42</span></article>
<article class="product" data-sku="camp-mug"><h2>Camp Mug</h2><a class="detail" href="/detail/camp-mug">Detail</a><span class="price">$12</span></article>
</main></body></html>`, cookieValue)
		case "/catalog-redesign":
			fmt.Fprint(w, `<!doctype html><html><body><main>
<section class="card" data-sku="trail-kit"><h2>Trail Kit</h2><strong class="cost">$42</strong></section>
</main></body></html>`)
		case "/detail/trail-kit":
			fmt.Fprint(w, `<!doctype html><html><body><main><h1>Trail Kit Detail</h1><p>Ready for camp.</p></main></body></html>`)
		case "/detail/camp-mug":
			fmt.Fprint(w, `<!doctype html><html><body><main><h1>Camp Mug Detail</h1><p>Enamel mug.</p></main></body></html>`)
		case "/order":
			if r.Method != http.MethodPost {
				t.Fatalf("order method = %s, want POST", r.Method)
			}
			username, password, ok := r.BasicAuth()
			if !ok || username != "agent" || password != "secret" {
				t.Fatalf("order auth = %q/%q ok=%v", username, password, ok)
			}
			cookie, err := r.Cookie("sid")
			if err != nil || cookie.Value != "complex-session" {
				t.Fatalf("order cookie = %#v err=%v", cookie, err)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read order body: %v", err)
			}
			orderBodies = append(orderBodies, string(body))
			fmt.Fprintf(w, `<html><body><span class="accepted">%s</span></body></html>`, body)
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	storePath := filepath.Join(t.TempDir(), "complex-adaptive.json")
	store, err := goscrapling.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	session, err := fetchers.NewFetcherSession(fetchers.FetcherSessionOptions{
		Client:  server.Client(),
		Headers: http.Header{"X-Complex-E2E": []string{"session"}},
		Store:   store,
	})
	if err != nil {
		t.Fatalf("NewFetcherSession: %v", err)
	}
	if _, err := session.Get(server.URL+"/login", fetchers.RequestOptions{}); err != nil {
		t.Fatalf("login: %v", err)
	}
	catalog, err := session.Get(server.URL+"/catalog", fetchers.RequestOptions{Params: url.Values{"segment": []string{"gear"}}})
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if got := catalog.CSS("main::attr(data-cookie)").Text(); got != "complex-session" {
		t.Fatalf("catalog session cookie = %q", got)
	}
	if got := strings.TrimSpace(catalog.CSS(".product::attr(data-sku)").Text()); got != "trail-kit\ncamp-mug" {
		t.Fatalf("catalog skus = %q", got)
	}
	order, err := session.Post(server.URL+"/order", fetchers.RequestOptions{
		Auth: &fetchers.BasicAuth{Username: "agent", Password: "secret"},
		Body: strings.NewReader("sku=trail-kit&qty=1"),
	})
	if err != nil {
		t.Fatalf("order: %v", err)
	}
	if got := order.CSS(".accepted::text").Text(); got != "sku=trail-kit&qty=1" {
		t.Fatalf("order echo = %q", got)
	}

	concurrent := fetchers.NewConcurrentFetcher(fetchers.ConcurrentFetcherOptions{Session: session, MaxConcurrency: 2})
	results := concurrent.Fetch(ctx, []fetchers.ConcurrentRequest{{URL: server.URL + "/detail/trail-kit"}, {URL: server.URL + "/detail/camp-mug"}})
	if len(results) != 2 {
		t.Fatalf("concurrent results = %d, want 2", len(results))
	}
	var titles []string
	for _, result := range results {
		if result.Err != nil {
			t.Fatalf("concurrent fetch: %v", result.Err)
		}
		titles = append(titles, result.Response.CSS("h1::text").Text())
	}
	sort.Strings(titles)
	if !reflect.DeepEqual(titles, []string{"Camp Mug Detail", "Trail Kit Detail"}) {
		t.Fatalf("concurrent titles = %#v", titles)
	}

	doc, err := goscrapling.Parse(bytes.NewReader(catalog.Body()), goscrapling.ParseOptions{URL: server.URL + "/catalog", Store: store})
	if err != nil {
		t.Fatalf("Parse catalog: %v", err)
	}
	if _, err := doc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "complex-price", AutoSave: true}); err != nil {
		t.Fatalf("save adaptive price: %v", err)
	}
	reloaded, err := goscrapling.NewFileStore(storePath)
	if err != nil {
		t.Fatalf("reload store: %v", err)
	}
	redesigned, err := session.Get(server.URL+"/catalog-redesign", fetchers.RequestOptions{Store: reloaded})
	if err != nil {
		t.Fatalf("redesigned: %v", err)
	}
	redesignedDoc, err := goscrapling.Parse(bytes.NewReader(redesigned.Body()), goscrapling.ParseOptions{URL: server.URL + "/catalog", Store: reloaded})
	if err != nil {
		t.Fatalf("Parse redesign: %v", err)
	}
	relocated, err := redesignedDoc.SelectCSS(ctx, ".price", goscrapling.SelectorOptions{Identifier: "complex-price", Adaptive: true, Percentage: 50})
	if err != nil {
		t.Fatalf("relocate price: %v", err)
	}
	if got := strings.TrimSpace(relocated.Text()); got != "$42" {
		t.Fatalf("relocated price = %q", got)
	}

	manager := spiders.NewSessionManager()
	if err := manager.Add("static", spiders.NewStaticSessionAdapter(session, spiders.StaticSessionAdapterOptions{}), spiders.SessionOptions{Default: true}); err != nil {
		t.Fatalf("add spider session: %v", err)
	}
	parseDetail := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		return []spiders.Output{spiders.Item(map[string]any{"title": response.CSS("h1::text").Text(), "sku": response.Meta["sku"]})}, nil
	}
	parseCatalog := func(_ context.Context, response spiders.Response) ([]spiders.Output, error) {
		trail, err := response.Follow("/detail/trail-kit", spiders.FollowOptions{Callback: parseDetail, Meta: map[string]any{"sku": "trail-kit"}})
		if err != nil {
			return nil, err
		}
		mug, err := response.Follow("/detail/camp-mug", spiders.FollowOptions{Callback: parseDetail, Meta: map[string]any{"sku": "camp-mug"}})
		if err != nil {
			return nil, err
		}
		return []spiders.Output{spiders.Next(trail), spiders.Next(mug)}, nil
	}
	crawl, err := (spiders.Crawler{Sessions: manager, DefaultCallback: parseCatalog}).Run(ctx, []spiders.Request{{URL: server.URL + "/catalog?segment=gear"}})
	if err != nil {
		t.Fatalf("spider crawl: %v", err)
	}
	var spiderTitles []string
	for _, item := range crawl.Items {
		spiderTitles = append(spiderTitles, fmt.Sprintf("%s:%s", item["sku"], item["title"]))
	}
	sort.Strings(spiderTitles)
	if !reflect.DeepEqual(spiderTitles, []string{"camp-mug:Camp Mug Detail", "trail-kit:Trail Kit Detail"}) {
		t.Fatalf("spider titles = %#v", spiderTitles)
	}

	binary := clitest.BuildBinary(t)
	outputPath := filepath.Join(t.TempDir(), "skus.txt")
	result := clitest.RunBinary(t, binary,
		"extract", "get", server.URL+"/catalog?segment=gear", outputPath,
		"--css-selector", ".product::attr(data-sku)",
	)
	if result.Err != nil {
		t.Fatalf("cli extract: %v\nstdout: %s\nstderr: %s", result.Err, result.Stdout, result.Stderr)
	}
	body, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read CLI output: %v", err)
	}
	if got := strings.TrimSpace(string(body)); got != "trail-kit\ncamp-mug" {
		t.Fatalf("CLI skus = %q", got)
	}
	if len(orderBodies) != 1 || orderBodies[0] != "sku=trail-kit&qty=1" {
		t.Fatalf("order bodies = %#v", orderBodies)
	}
}
