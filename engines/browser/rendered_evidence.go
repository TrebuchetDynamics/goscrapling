package browser

import (
	"bytes"
	"encoding/json"
	"strings"

	"golang.org/x/net/html"
)

type RenderedEvidence struct {
	root   *html.Node
	labels map[string]string
}

type Link struct {
	Text string
	Href string
}

type StructuredData struct {
	OpenGraph map[string]string
	JSONLD    []map[string]any
}

func NewRenderedEvidence(body []byte) (RenderedEvidence, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return RenderedEvidence{}, err
	}
	labels := map[string]string{}
	collectLabels(root, labels)
	return RenderedEvidence{root: root, labels: labels}, nil
}

func (e RenderedEvidence) Markdown() string {
	var renderer markdownRenderer
	renderer.renderBlocks(e.root)
	return strings.TrimSpace(renderer.output.String())
}

func (e RenderedEvidence) SemanticNodes() []SemanticNode {
	builder := semanticBuilder{labels: e.labels}
	builder.walk(e.root)
	return builder.nodes
}

func (e RenderedEvidence) InteractiveNodes() []SemanticNode {
	nodes := e.SemanticNodes()
	interactive := make([]SemanticNode, 0, len(nodes))
	for _, node := range nodes {
		if node.Interactive {
			interactive = append(interactive, node)
		}
	}
	return interactive
}

func (e RenderedEvidence) Links() []Link {
	var links []Link
	walkHTML(e.root, func(node *html.Node) {
		if node.Type != html.ElementNode || !strings.EqualFold(node.Data, "a") {
			return
		}
		href := attrValue(node, "href")
		text := normalizeMarkdownSpace(nodeVisibleText(node))
		if href != "" && text != "" {
			links = append(links, Link{Text: text, Href: href})
		}
	})
	return links
}

func (e RenderedEvidence) StructuredData() StructuredData {
	data := StructuredData{OpenGraph: map[string]string{}}
	walkHTML(e.root, func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		switch strings.ToLower(node.Data) {
		case "meta":
			property := attrValue(node, "property")
			if strings.HasPrefix(property, "og:") {
				data.OpenGraph[strings.TrimPrefix(property, "og:")] = attrValue(node, "content")
			}
		case "script":
			if !strings.EqualFold(attrValue(node, "type"), "application/ld+json") {
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
