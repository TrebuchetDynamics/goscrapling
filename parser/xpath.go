package parser

import (
	"fmt"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

func (d *Document) XPath(expr string) Selection {
	if d == nil || d.root == nil {
		return Selection{}
	}

	selection, err := selectXPath(d, d.root, expr)
	if err != nil {
		return Selection{}
	}
	return selection
}

func (e *Element) XPath(expr string) Selection {
	if e == nil || e.node == nil {
		return Selection{}
	}

	selection, err := selectXPath(e.doc, e.node, expr)
	if err != nil {
		return Selection{}
	}
	return selection
}

func (s Selection) XPath(expr string) Selection {
	var combined Selection
	for _, element := range s.elements {
		if element == nil || element.node == nil {
			continue
		}
		selection, err := selectXPath(element.doc, element.node, expr)
		if err != nil {
			return Selection{}
		}
		combined = combineSelections(combined, selection)
	}
	return combined
}

func selectXPath(doc *Document, root *html.Node, expr string) (Selection, error) {
	nodes, err := htmlquery.QueryAll(root, expr)
	if err != nil {
		return Selection{}, fmt.Errorf("%w %q: %v", ErrInvalidSelector, expr, err)
	}

	elements := make([]*Element, 0, len(nodes))
	values := make(TextHandlers, 0)
	valueSources := make([]*Element, 0)
	for _, node := range nodes {
		if node == nil {
			continue
		}
		value, ok := xpathNodeValue(node)
		if !ok && node.Type == html.ElementNode {
			elements = append(elements, &Element{doc: doc, node: node})
			continue
		}
		if value == "" {
			continue
		}
		values = append(values, TextHandler(value))
		valueSources = append(valueSources, &Element{doc: doc, node: node})
	}

	if len(values) > 0 {
		return Selection{elements: valueSources, values: values, extract: true}, nil
	}
	return Selection{elements: elements}, nil
}

func xpathNodeValue(node *html.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Type == html.TextNode {
		return normalizeSpace(node.Data), true
	}
	if isHTMLQueryAttributeNode(node) {
		return htmlquery.InnerText(node), true
	}
	return "", false
}

func isHTMLQueryAttributeNode(node *html.Node) bool {
	return node != nil &&
		node.Type == html.ElementNode &&
		node.Parent == nil &&
		node.FirstChild != nil &&
		node.FirstChild.Type == html.TextNode
}
