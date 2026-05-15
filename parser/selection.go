package parser

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

type Selection struct {
	elements []*Element
	values   TextHandlers
	extract  bool
}

func (s Selection) Len() int {
	if s.extract {
		return len(s.values)
	}
	return len(s.elements)
}

func (s Selection) First() (*Element, bool) {
	if len(s.elements) == 0 {
		return nil, false
	}
	return s.elements[0], true
}

func (s Selection) Text() string {
	if s.extract {
		return strings.Join(s.values.Strings(), "\n")
	}

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
	if s.extract {
		return strings.Join(s.values.Strings(), "\n"), nil
	}

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

func (s Selection) Get(defaultValue ...string) TextHandler {
	return s.GetAll().Get(defaultValue...)
}

func (s Selection) GetAll() TextHandlers {
	if s.extract {
		return s.values.GetAll()
	}

	values := make(TextHandlers, 0, len(s.elements))
	for _, element := range s.elements {
		value, err := element.HTML()
		if err != nil || value == "" {
			continue
		}
		values = append(values, TextHandler(value))
	}
	return values
}

func (s Selection) Extract() TextHandlers {
	return s.GetAll()
}

func (s Selection) Regex(pattern string) (TextHandlers, error) {
	return s.GetAll().Regex(pattern)
}

func (s Selection) RegexFirst(pattern string, defaultValue ...string) (TextHandler, error) {
	return s.GetAll().RegexFirst(pattern, defaultValue...)
}

func (s Selection) JSON() (any, error) {
	return s.Get().JSON()
}

func (s Selection) withTextValues() Selection {
	values := make(TextHandlers, 0, len(s.elements))
	sources := make([]*Element, 0, len(s.elements))

	for _, element := range s.elements {
		if element == nil || element.node == nil {
			continue
		}
		for child := element.node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != html.TextNode {
				continue
			}
			text := normalizeSpace(child.Data)
			if text == "" {
				continue
			}
			values = append(values, TextHandler(text))
			sources = append(sources, element)
		}
	}

	return Selection{elements: sources, values: values, extract: true}
}

func (s Selection) withAttrValues(name string) Selection {
	values := make(TextHandlers, 0, len(s.elements))
	sources := make([]*Element, 0, len(s.elements))

	for _, element := range s.elements {
		value, ok := element.Attr(name)
		if !ok {
			continue
		}
		values = append(values, TextHandler(value))
		sources = append(sources, element)
	}

	return Selection{elements: sources, values: values, extract: true}
}

func combineSelections(left Selection, right Selection) Selection {
	elements := make([]*Element, 0, len(left.elements)+len(right.elements))
	elements = append(elements, left.elements...)
	elements = append(elements, right.elements...)

	if left.extract || right.extract {
		values := make(TextHandlers, 0, left.Len()+right.Len())
		values = append(values, left.GetAll()...)
		values = append(values, right.GetAll()...)
		return Selection{elements: elements, values: values, extract: true}
	}

	return Selection{elements: elements}
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

func (e *Element) Attrs() AttributesHandler {
	if e == nil || e.node == nil {
		return newAttributesHandler(nil)
	}

	attributes := make(map[string]string, len(e.node.Attr))
	for _, attr := range e.node.Attr {
		attributes[strings.ToLower(attr.Key)] = attr.Val
	}
	return newAttributesHandler(attributes)
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
