package goscrapling

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Selection struct {
	elements []*Element
}

func (s Selection) Len() int {
	return len(s.elements)
}

func (s Selection) First() (*Element, bool) {
	if len(s.elements) == 0 {
		return nil, false
	}
	return s.elements[0], true
}

func (s Selection) Text() string {
	parts := make([]string, 0, len(s.elements))
	for _, element := range s.elements {
		text := element.Text()
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func (s Selection) HTML() (string, error) {
	parts := make([]string, 0, len(s.elements))
	for _, element := range s.elements {
		body, err := element.HTML()
		if err != nil {
			return "", err
		}
		if body != "" {
			parts = append(parts, body)
		}
	}
	return strings.Join(parts, "\n"), nil
}

type Element struct {
	doc  *Document
	node *html.Node
}

func (e *Element) TagName() string {
	if e == nil || e.node == nil {
		return ""
	}
	return strings.ToLower(e.node.Data)
}

func (e *Element) Text() string {
	if e == nil || e.node == nil {
		return ""
	}
	return nodeText(e.node)
}

func (e *Element) Attr(name string) (string, bool) {
	if e == nil || e.node == nil {
		return "", false
	}

	name = strings.ToLower(name)
	for _, attr := range e.node.Attr {
		if strings.ToLower(attr.Key) == name {
			return attr.Val, true
		}
	}

	return "", false
}

func (e *Element) HTML() (string, error) {
	if e == nil || e.node == nil {
		return "", nil
	}
	var output strings.Builder
	if err := html.Render(&output, e.node); err != nil {
		return "", err
	}
	return output.String(), nil
}

func normalizeSpace(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}

func nodeText(node *html.Node) string {
	if node == nil {
		return ""
	}

	var parts []string
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			text := normalizeSpace(current.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)

	return strings.Join(parts, " ")
}
