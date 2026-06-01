package browsertools

import (
	"context"
	"net/http"
	"testing"

	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
)

func TestGormesBrowserExtractionTools(t *testing.T) {
	adapter := BrowserExtractionAdapter{Engine: fakeBrowserEngine{body: []byte(`<!doctype html>
<html><head>
<meta property="og:title" content="Agent Browser Fixture">
<script type="application/ld+json">{"@type":"Article","headline":"Agent Browser Fixture"}</script>
</head><body>
<main>
  <h1>Agent Browser Fixture</h1>
  <p>Read the <a href="/docs">docs</a>.</p>
  <form action="/signup"><label for="email">Email</label><input id="email" name="email" value="juan@example.com"></form>
  <button>Continue</button>
</main>
</body></html>`)}}

	cases := []struct {
		tool  string
		check func(*testing.T, ToolResult)
	}{
		{
			tool: ToolRenderedMarkdown,
			check: func(t *testing.T, result ToolResult) {
				if result.Markdown == "" || result.Markdown != "# Agent Browser Fixture\n\nRead the [docs](/docs)." {
					t.Fatalf("markdown = %q", result.Markdown)
				}
			},
		},
		{
			tool: ToolLinks,
			check: func(t *testing.T, result ToolResult) {
				if len(result.Links) != 1 || result.Links[0].Text != "docs" || result.Links[0].Href != "/docs" {
					t.Fatalf("links = %#v", result.Links)
				}
			},
		},
		{
			tool: ToolSemanticTree,
			check: func(t *testing.T, result ToolResult) {
				if len(result.SemanticTree) == 0 || !hasRole(result.SemanticTree, "button", "Continue") || !hasRole(result.SemanticTree, "textbox", "Email") {
					t.Fatalf("semantic tree = %#v", result.SemanticTree)
				}
			},
		},
		{
			tool: ToolStructuredData,
			check: func(t *testing.T, result ToolResult) {
				if result.StructuredData.OpenGraph["title"] != "Agent Browser Fixture" || len(result.StructuredData.JSONLD) != 1 {
					t.Fatalf("structured data = %#v", result.StructuredData)
				}
			},
		},
		{
			tool: ToolInteractiveElements,
			check: func(t *testing.T, result ToolResult) {
				if len(result.InteractiveElements) != 3 {
					t.Fatalf("interactive elements = %#v", result.InteractiveElements)
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.tool, func(t *testing.T) {
			result, err := adapter.Call(context.Background(), ToolCall{Tool: tt.tool, URL: "https://example.com/app"})
			if err != nil {
				t.Fatalf("Call returned error: %v", err)
			}
			if result.URL != "https://example.com/rendered" || result.StatusCode != http.StatusOK || result.Tool != tt.tool {
				t.Fatalf("metadata = %#v", result)
			}
			tt.check(t, result)
		})
	}
}

type fakeBrowserEngine struct{ body []byte }

func (e fakeBrowserEngine) Fetch(context.Context, browser.BrowserRequest) (browser.BrowserResult, error) {
	return browser.BrowserResult{URL: "https://example.com/rendered", StatusCode: http.StatusOK, Body: e.body}, nil
}

func hasRole(nodes []browser.SemanticNode, role, name string) bool {
	for _, node := range nodes {
		if node.Role == role && node.Name == name {
			return true
		}
	}
	return false
}
