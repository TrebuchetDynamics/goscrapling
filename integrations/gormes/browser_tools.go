package gormes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"golang.org/x/net/html"
)

const (
	ToolRenderedMarkdown    = "rendered_markdown"
	ToolLinks               = "links"
	ToolSemanticTree        = "semantic_tree"
	ToolStructuredData      = "structured_data"
	ToolInteractiveElements = "interactive_elements"
)

var ErrUnknownTool = errors.New("unknown gormes browser extraction tool")

type BrowserExtractionAdapter struct {
	Engine browser.BrowserEngine
}

type ToolCall struct {
	Tool string
	URL  string
	Opts browser.BrowserOptions
}

type ToolResult struct {
	Tool                string
	URL                 string
	StatusCode          int
	Markdown            string
	Links               []Link
	SemanticTree        []browser.SemanticNode
	StructuredData      StructuredData
	InteractiveElements []browser.SemanticNode
}

type Link struct {
	Text string
	Href string
}

type StructuredData struct {
	OpenGraph map[string]string
	JSONLD    []map[string]any
}

func (a BrowserExtractionAdapter) Call(ctx context.Context, call ToolCall) (ToolResult, error) {
	fetcher := browser.BrowserFetcher{Engine: a.Engine}
	if call.Opts.Headers == nil {
		call.Opts.Headers = http.Header{}
	}

	switch call.Tool {
	case ToolRenderedMarkdown:
		dump, err := fetcher.FetchMarkdown(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Tool: call.Tool, URL: dump.URL, StatusCode: dump.StatusCode, Markdown: dump.Markdown}, nil
	case ToolSemanticTree, ToolInteractiveElements:
		tree, err := fetcher.FetchSemanticTree(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		result := ToolResult{Tool: call.Tool, URL: tree.URL, StatusCode: tree.StatusCode, SemanticTree: tree.Nodes}
		if call.Tool == ToolInteractiveElements {
			result.SemanticTree = nil
			for _, node := range tree.Nodes {
				if node.Interactive {
					result.InteractiveElements = append(result.InteractiveElements, node)
				}
			}
		}
		return result, nil
	case ToolLinks, ToolStructuredData:
		response, err := fetcher.Fetch(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		result := ToolResult{Tool: call.Tool, URL: response.URL(), StatusCode: response.StatusCode()}
		if call.Tool == ToolLinks {
			result.Links = extractLinks(response.Body())
		} else {
			result.StructuredData = extractStructuredData(response.Body())
		}
		return result, nil
	default:
		return ToolResult{}, ErrUnknownTool
	}
}

func extractLinks(body []byte) []Link {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}
	var links []Link
	walkHTML(root, func(node *html.Node) {
		if node.Type != html.ElementNode || !strings.EqualFold(node.Data, "a") {
			return
		}
		href := htmlAttr(node, "href")
		text := normalizeText(htmlText(node))
		if href != "" && text != "" {
			links = append(links, Link{Text: text, Href: href})
		}
	})
	return links
}

func extractStructuredData(body []byte) StructuredData {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return StructuredData{}
	}
	data := StructuredData{OpenGraph: map[string]string{}}
	walkHTML(root, func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		switch strings.ToLower(node.Data) {
		case "meta":
			property := htmlAttr(node, "property")
			if strings.HasPrefix(property, "og:") {
				data.OpenGraph[strings.TrimPrefix(property, "og:")] = htmlAttr(node, "content")
			}
		case "script":
			if !strings.EqualFold(htmlAttr(node, "type"), "application/ld+json") {
				return
			}
			var value map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(htmlText(node))), &value); err == nil {
				data.JSONLD = append(data.JSONLD, value)
			}
		}
	})
	if len(data.OpenGraph) == 0 {
		data.OpenGraph = nil
	}
	return data
}

func walkHTML(node *html.Node, visit func(*html.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkHTML(child, visit)
	}
}

func htmlAttr(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func htmlText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if text := htmlText(child); strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
