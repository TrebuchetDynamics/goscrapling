package mcp

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/fetchers"
)

func TestMCPServerOpenSessionRejectsDuplicateWhileFirstCreateIsInFlight(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	factory := BrowserFactory(func(context.Context, SessionType, browser.BrowserOptions, int) (BrowserSession, error) {
		entered <- struct{}{}
		<-release
		return &fakeBrowserSession{alive: true}, nil
	})
	server := NewServer(ServerOptions{BrowserFactory: factory})

	firstDone := make(chan error, 1)
	go func() {
		_, err := server.OpenSession(context.Background(), OpenSessionRequest{SessionID: "shared"})
		firstDone <- err
	}()
	<-entered

	secondDone := make(chan error, 1)
	go func() {
		_, err := server.OpenSession(context.Background(), OpenSessionRequest{SessionID: "shared"})
		secondDone <- err
	}()

	select {
	case err := <-secondDone:
		if err == nil || !strings.Contains(err.Error(), `session "shared" already exists`) {
			t.Fatalf("second OpenSession error = %v, want duplicate-session error", err)
		}
	case <-time.After(50 * time.Millisecond):
		close(release)
		<-firstDone
		t.Fatal("second OpenSession blocked behind duplicate in-flight session creation")
	}

	close(release)
	if err := <-firstDone; err != nil {
		t.Fatalf("first OpenSession returned error: %v", err)
	}
	select {
	case <-entered:
		t.Fatal("duplicate OpenSession called browser factory")
	default:
	}
}

func TestMCPServerTools(t *testing.T) {
	static := &fakeStaticClient{pages: map[string]string{
		"https://example.com/catalog": `<!doctype html><html><body><nav>Skip nav</nav><main><article class="product">Alpha Tool</article></main></body></html>`,
		"https://example.com/about":   `<!doctype html><html><body><main>About page</main></body></html>`,
	}}
	factory := &fakeBrowserFactory{}
	server := NewServer(ServerOptions{Static: static, BrowserFactory: factory.New})

	t.Run("tool schemas include deterministic Scrapling MCP surface", func(t *testing.T) {
		tools := server.Tools()
		gotNames := make([]string, 0, len(tools))
		for _, tool := range tools {
			gotNames = append(gotNames, tool.Name)
		}
		sort.Strings(gotNames)
		wantNames := []string{ToolBulkFetch, ToolBulkGet, ToolBulkStealthyFetch, ToolCloseSession, ToolFetch, ToolGet, ToolListSessions, ToolOpenSession, ToolScreenshot, ToolStealthyFetch}
		sort.Strings(wantNames)
		if !reflect.DeepEqual(gotNames, wantNames) {
			t.Fatalf("tool names = %#v, want %#v", gotNames, wantNames)
		}
		getSchema := toolByName(t, tools, ToolGet).InputSchema
		for _, field := range []string{"url", "css_selector", "main_content_only", "extraction_type"} {
			if _, ok := getSchema.Properties[field]; !ok {
				t.Fatalf("get schema missing field %q: %#v", field, getSchema.Properties)
			}
		}
		if !toolByName(t, tools, ToolScreenshot).ReturnsImage {
			t.Fatal("screenshot tool should advertise image content")
		}
	})

	t.Run("get and bulk_get use static seam with CSS and main content extraction", func(t *testing.T) {
		result, err := server.Get(context.Background(), GetRequest{URL: "https://example.com/catalog", CSSSelector: "article.product", ExtractionType: ExtractionText})
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if result.Status != http.StatusOK || result.URL != "https://example.com/catalog" || !reflect.DeepEqual(result.Content, []string{"Alpha Tool"}) {
			t.Fatalf("get result = %#v", result)
		}
		bulk, err := server.BulkGet(context.Background(), BulkGetRequest{URLs: []string{"https://example.com/catalog", "https://example.com/about"}, MainContentOnly: true, ExtractionType: ExtractionText})
		if err != nil {
			t.Fatalf("BulkGet returned error: %v", err)
		}
		if len(bulk) != 2 || strings.Contains(bulk[0].Content[0], "Skip nav") || bulk[1].Content[0] != "About page" {
			t.Fatalf("bulk get = %#v", bulk)
		}
		if got := static.Calls(); !reflect.DeepEqual(got, []string{"https://example.com/catalog", "https://example.com/catalog", "https://example.com/about"}) {
			t.Fatalf("static calls = %#v", got)
		}
	})

	t.Run("browser, stealth, screenshot, and session tools use fake browser seam", func(t *testing.T) {
		created, err := server.OpenSession(context.Background(), OpenSessionRequest{
			SessionType: SessionStealthy,
			SessionID:   "agent-stealth",
			Headless:    true,
			MaxPages:    3,
			Stealth:     browser.BrowserStealthOptions{Enabled: true, GenerateHeaders: true, HideCanvas: true},
		})
		if err != nil {
			t.Fatalf("OpenSession returned error: %v", err)
		}
		if created.SessionID != "agent-stealth" || created.SessionType != SessionStealthy || !created.IsAlive || !strings.Contains(created.Message, "created") {
			t.Fatalf("created session = %#v", created)
		}
		infos, err := server.ListSessions(context.Background())
		if err != nil {
			t.Fatalf("ListSessions returned error: %v", err)
		}
		if len(infos) != 1 || infos[0].SessionID != "agent-stealth" || infos[0].SessionType != SessionStealthy || !infos[0].IsAlive {
			t.Fatalf("sessions = %#v", infos)
		}

		fetchResult, err := server.Fetch(context.Background(), FetchRequest{URL: "https://example.com/app", ExtractionType: ExtractionText, Wait: 10 * time.Millisecond, WaitSelector: "#ready"})
		if err != nil {
			t.Fatalf("Fetch returned error: %v", err)
		}
		stealthResult, err := server.StealthyFetch(context.Background(), FetchRequest{URL: "https://example.com/protected", SessionID: "agent-stealth", ExtractionType: ExtractionText})
		if err != nil {
			t.Fatalf("StealthyFetch returned error: %v", err)
		}
		bulkFetch, err := server.BulkFetch(context.Background(), BulkFetchRequest{URLs: []string{"https://example.com/a", "https://example.com/b"}, ExtractionType: ExtractionText})
		if err != nil {
			t.Fatalf("BulkFetch returned error: %v", err)
		}
		bulkStealth, err := server.BulkStealthyFetch(context.Background(), BulkFetchRequest{URLs: []string{"https://example.com/s1", "https://example.com/s2"}, SessionID: "agent-stealth", ExtractionType: ExtractionText})
		if err != nil {
			t.Fatalf("BulkStealthyFetch returned error: %v", err)
		}
		if fetchResult.Content[0] != "dynamic:https://example.com/app" || stealthResult.Content[0] != "stealthy:https://example.com/protected" || len(bulkFetch) != 2 || len(bulkStealth) != 2 {
			t.Fatalf("browser results = %#v / %#v / %#v / %#v", fetchResult, stealthResult, bulkFetch, bulkStealth)
		}

		blocks, err := server.Screenshot(context.Background(), ScreenshotRequest{URL: "https://example.com/image", SessionID: "agent-stealth", ImageType: "jpeg", Quality: 80})
		if err != nil {
			t.Fatalf("Screenshot returned error: %v", err)
		}
		if len(blocks) != 2 || blocks[0].Type != ContentImage || blocks[0].MimeType != "image/jpeg" || string(blocks[0].Data) != "shot:https://example.com/image" || blocks[1].Text != "https://example.com/image" {
			t.Fatalf("screenshot blocks = %#v", blocks)
		}

		closed, err := server.CloseSession(context.Background(), CloseSessionRequest{SessionID: "agent-stealth"})
		if err != nil {
			t.Fatalf("CloseSession returned error: %v", err)
		}
		if closed.SessionID != "agent-stealth" || !strings.Contains(closed.Message, "closed") {
			t.Fatalf("closed session = %#v", closed)
		}
		infos, err = server.ListSessions(context.Background())
		if err != nil {
			t.Fatalf("ListSessions after close returned error: %v", err)
		}
		if len(infos) != 0 {
			t.Fatalf("sessions after close = %#v", infos)
		}

		seen := factory.Requests()
		if !hasBrowserCall(seen, SessionDynamic, "https://example.com/app") || !hasBrowserCall(seen, SessionStealthy, "https://example.com/protected") {
			t.Fatalf("browser calls = %#v", seen)
		}
	})
}

func toolByName(t *testing.T, tools []ToolSpec, name string) ToolSpec {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found in %#v", name, tools)
	return ToolSpec{}
}

type fakeStaticClient struct {
	mu    sync.Mutex
	pages map[string]string
	calls []string
}

func (c *fakeStaticClient) Get(ctx context.Context, rawURL string, opts fetchers.RequestOptions) (*goscrapling.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.calls = append(c.calls, rawURL)
	c.mu.Unlock()
	body := c.pages[rawURL]
	if body == "" {
		body = "<html><body><main>missing</main></body></html>"
	}
	return goscrapling.NewResponse(strings.NewReader(body), goscrapling.ResponseOptions{
		URL:        rawURL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request: goscrapling.RequestMetadata{
			Method:  http.MethodGet,
			URL:     rawURL,
			Headers: opts.Headers,
		},
	})
}

func (c *fakeStaticClient) Calls() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	calls := append([]string(nil), c.calls...)
	return calls
}

type browserCall struct {
	SessionType SessionType
	URL         string
	Options     browser.BrowserOptions
}

type fakeBrowserFactory struct {
	mu       sync.Mutex
	requests []browserCall
}

func (f *fakeBrowserFactory) New(_ context.Context, sessionType SessionType, opts browser.BrowserOptions, _ int) (BrowserSession, error) {
	return &fakeBrowserSession{factory: f, sessionType: sessionType, defaults: opts, alive: true}, nil
}

func (f *fakeBrowserFactory) record(call browserCall) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, call)
}

func (f *fakeBrowserFactory) Requests() []browserCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := append([]browserCall(nil), f.requests...)
	return out
}

type fakeBrowserSession struct {
	factory     *fakeBrowserFactory
	sessionType SessionType
	defaults    browser.BrowserOptions
	alive       bool
}

func (s *fakeBrowserSession) Fetch(ctx context.Context, rawURL string, opts browser.BrowserOptions) (*goscrapling.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.factory.record(browserCall{SessionType: s.sessionType, URL: rawURL, Options: opts})
	return goscrapling.NewResponse(strings.NewReader(fmt.Sprintf("<html><body><main>%s:%s</main></body></html>", s.sessionType, rawURL)), goscrapling.ResponseOptions{
		URL:        rawURL,
		StatusCode: http.StatusOK,
		Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
	})
}

func (s *fakeBrowserSession) Screenshot(ctx context.Context, rawURL string, opts browser.BrowserOptions) ([]byte, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	s.factory.record(browserCall{SessionType: s.sessionType, URL: rawURL, Options: opts})
	return []byte("shot:" + rawURL), rawURL, nil
}

func (s *fakeBrowserSession) Close(context.Context) error {
	s.alive = false
	return nil
}

func (s *fakeBrowserSession) Alive() bool { return s.alive }

func hasBrowserCall(calls []browserCall, sessionType SessionType, rawURL string) bool {
	for _, call := range calls {
		if call.SessionType == sessionType && call.URL == rawURL {
			return true
		}
	}
	return false
}
