package browser

import (
	"context"
	"strings"

	"golang.org/x/net/html"
)

type BrowserSemanticTree struct {
	URL        string
	StatusCode int
	Nodes      []SemanticNode
	Response   *Response
}

type SemanticNode struct {
	ID          int
	Role        string
	Name        string
	Text        string
	Tag         string
	XPath       string
	Interactive bool
	Disabled    bool
	Value       string
	Href        string
	Checked     bool
}

func (f BrowserFetcher) FetchSemanticTree(ctx context.Context, rawURL string, opts BrowserOptions) (BrowserSemanticTree, error) {
	response, err := f.Fetch(ctx, rawURL, opts)
	if err != nil {
		return BrowserSemanticTree{}, err
	}
	nodes, err := HTMLSemanticNodes(response.Body())
	if err != nil {
		return BrowserSemanticTree{}, err
	}
	return BrowserSemanticTree{
		URL:        response.URL(),
		StatusCode: response.StatusCode(),
		Nodes:      nodes,
		Response:   response,
	}, nil
}

func HTMLSemanticNodes(body []byte) ([]SemanticNode, error) {
	evidence, err := NewRenderedEvidence(body)
	if err != nil {
		return nil, err
	}
	return evidence.SemanticNodes(), nil
}

type semanticBuilder struct {
	labels map[string]string
	nodes  []SemanticNode
}

func (b *semanticBuilder) walk(node *html.Node) {
	if node == nil || skipSemanticNode(node) {
		return
	}
	if node.Type == html.ElementNode {
		if semantic, ok := b.nodeForElement(node); ok {
			semantic.ID = len(b.nodes) + 1
			semantic.XPath = nodeXPath(node)
			b.nodes = append(b.nodes, semantic)
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		b.walk(child)
	}
}

func (b *semanticBuilder) nodeForElement(node *html.Node) (SemanticNode, bool) {
	tag := strings.ToLower(node.Data)
	text := normalizeMarkdownSpace(nodeVisibleText(node))
	n := SemanticNode{
		Tag:      tag,
		Text:     text,
		Name:     accessibleName(node, b.labels, text),
		Disabled: hasAttr(node, "disabled"),
	}

	switch tag {
	case "a":
		href := attrValue(node, "href")
		if href == "" || n.Name == "" {
			return SemanticNode{}, false
		}
		n.Role = "link"
		n.Href = href
		n.Interactive = true
	case "button":
		if n.Name == "" {
			return SemanticNode{}, false
		}
		n.Role = "button"
		n.Interactive = true
	case "input":
		typeValue := strings.ToLower(attrValue(node, "type"))
		if typeValue == "" {
			typeValue = "text"
		}
		n.Value = attrValue(node, "value")
		n.Checked = hasAttr(node, "checked")
		n.Interactive = true
		switch typeValue {
		case "checkbox":
			n.Role = "checkbox"
		case "radio":
			n.Role = "radio"
		case "submit", "button":
			n.Role = "button"
			if n.Name == "" {
				n.Name = n.Value
			}
		default:
			n.Role = "textbox"
		}
	case "textarea":
		n.Role = "textbox"
		n.Value = text
		n.Interactive = true
	case "select":
		n.Role = "combobox"
		n.Interactive = true
	case "h1", "h2", "h3", "h4", "h5", "h6":
		if n.Name == "" {
			return SemanticNode{}, false
		}
		n.Role = "heading"
	default:
		return SemanticNode{}, false
	}
	return n, true
}

func collectLabels(node *html.Node, labels map[string]string) {
	if node == nil || skipSemanticNode(node) {
		return
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, "label") {
		if id := attrValue(node, "for"); id != "" {
			labels[id] = normalizeMarkdownSpace(nodeVisibleText(node))
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectLabels(child, labels)
	}
}

func accessibleName(node *html.Node, labels map[string]string, text string) string {
	for _, attr := range []string{"aria-label", "title", "alt", "placeholder"} {
		if value := normalizeMarkdownSpace(attrValue(node, attr)); value != "" {
			return value
		}
	}
	if id := attrValue(node, "id"); id != "" {
		if label := labels[id]; label != "" {
			return label
		}
	}
	if value := normalizeMarkdownSpace(attrValue(node, "value")); value != "" && strings.EqualFold(node.Data, "button") {
		return value
	}
	return text
}

func nodeVisibleText(node *html.Node) string {
	if node == nil || skipSemanticNode(node) {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	var parts []string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if text := nodeVisibleText(child); strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func skipSemanticNode(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "script", "style", "noscript", "template", "svg":
		return true
	default:
		return false
	}
}

func nodeXPath(node *html.Node) string {
	if node == nil || node.Type != html.ElementNode {
		return ""
	}
	var parts []string
	for current := node; current != nil; current = current.Parent {
		if current.Type != html.ElementNode {
			continue
		}
		parts = append([]string{strings.ToLower(current.Data) + "[" + siblingElementIndex(current) + "]"}, parts...)
	}
	return "/" + strings.Join(parts, "/")
}

func siblingElementIndex(node *html.Node) string {
	index := 1
	for sibling := node.PrevSibling; sibling != nil; sibling = sibling.PrevSibling {
		if sibling.Type == html.ElementNode && strings.EqualFold(sibling.Data, node.Data) {
			index++
		}
	}
	return strconvItoa(index)
}

func hasAttr(node *html.Node, name string) bool {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return true
		}
	}
	return false
}

func strconvItoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}
