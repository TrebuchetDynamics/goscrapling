package browser

import (
	"context"
	"net/http"
	"testing"
)

func TestBrowserSemanticTreeExtraction(t *testing.T) {
	engine := &recordingBrowserEngine{
		result: BrowserResult{
			URL:        "https://example.com/rendered",
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
			Body: []byte(`<!doctype html>
<html>
<body>
  <main>
    <h1>Checkout</h1>
    <a href="/cart">Cart</a>
    <button id="buy">Buy now</button>
    <label for="email">Email address</label>
    <input id="email" name="email" value="juan@example.com" />
    <input id="tos" type="checkbox" checked disabled />
    <script>ignored()</script>
    <style>button { color: red }</style>
  </main>
</body>
</html>`),
		},
	}
	fetcher := BrowserFetcher{Engine: engine}

	tree, err := fetcher.FetchSemanticTree(context.Background(), "https://example.com/app", BrowserOptions{LoadDOM: true})
	if err != nil {
		t.Fatalf("FetchSemanticTree returned error: %v", err)
	}
	if tree.URL != "https://example.com/rendered" || tree.StatusCode != http.StatusOK {
		t.Fatalf("tree metadata = url %q status %d", tree.URL, tree.StatusCode)
	}

	button := findSemanticNode(t, tree.Nodes, "button", "Buy now")
	if button.Role != "button" || !button.Interactive || button.XPath == "" {
		t.Fatalf("button node = %#v", button)
	}

	link := findSemanticNode(t, tree.Nodes, "link", "Cart")
	if link.Tag != "a" || link.Href != "/cart" || !link.Interactive {
		t.Fatalf("link node = %#v", link)
	}

	textbox := findSemanticNode(t, tree.Nodes, "textbox", "Email address")
	if textbox.Tag != "input" || textbox.Value != "juan@example.com" || !textbox.Interactive {
		t.Fatalf("textbox node = %#v", textbox)
	}

	checkbox := findSemanticNode(t, tree.Nodes, "checkbox", "")
	if !checkbox.Checked || !checkbox.Disabled || !checkbox.Interactive {
		t.Fatalf("checkbox node = %#v", checkbox)
	}

	for _, node := range tree.Nodes {
		if node.Tag == "script" || node.Tag == "style" || node.Name == "ignored()" || node.Name == "button { color: red }" {
			t.Fatalf("semantic tree kept noise node: %#v", node)
		}
	}
}

func findSemanticNode(t *testing.T, nodes []SemanticNode, role, name string) SemanticNode {
	t.Helper()
	for _, node := range nodes {
		if node.Role == role && node.Name == name {
			return node
		}
	}
	t.Fatalf("missing semantic node role=%q name=%q in %#v", role, name, nodes)
	return SemanticNode{}
}
