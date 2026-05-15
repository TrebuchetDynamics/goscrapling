package goscrapling

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

type TextSearchOptions struct {
	Partial       bool
	CaseSensitive bool
	CleanMatch    *bool
}

type SimilarOptions struct {
	Threshold        float64
	IgnoreAttributes []string
	MatchText        bool
}

func (s Selection) Search(predicate func(*Element) bool) (*Element, bool) {
	if predicate == nil {
		return nil, false
	}
	for _, element := range s.elements {
		if element != nil && predicate(element) {
			return element, true
		}
	}
	return nil, false
}

func (s Selection) Filter(predicate func(*Element) bool) Selection {
	if predicate == nil {
		return Selection{}
	}
	elements := make([]*Element, 0, len(s.elements))
	for _, element := range s.elements {
		if element != nil && predicate(element) {
			elements = append(elements, element)
		}
	}
	return Selection{elements: elements}
}

func (d *Document) FindByText(text string, opts TextSearchOptions) (*Element, bool) {
	return firstElement(d.FindAllByText(text, opts))
}

func (d *Document) FindAllByText(text string, opts TextSearchOptions) Selection {
	if d == nil || d.root == nil {
		return Selection{}
	}
	return findAllByText(d, d.root, text, opts)
}

func (e *Element) FindByText(text string, opts TextSearchOptions) (*Element, bool) {
	return firstElement(e.FindAllByText(text, opts))
}

func (e *Element) FindAllByText(text string, opts TextSearchOptions) Selection {
	if e == nil || e.node == nil {
		return Selection{}
	}
	return findAllByText(e.doc, e.node, text, opts)
}

func (d *Document) FindByRegex(pattern any, opts TextSearchOptions) (*Element, bool) {
	selection, err := d.FindAllByRegex(pattern, opts)
	if err != nil {
		return nil, false
	}
	return firstElement(selection)
}

func (d *Document) FindAllByRegex(pattern any, opts TextSearchOptions) (Selection, error) {
	if d == nil || d.root == nil {
		return Selection{}, nil
	}
	return findAllByRegex(d, d.root, pattern, opts)
}

func (e *Element) FindByRegex(pattern any, opts TextSearchOptions) (*Element, bool) {
	selection, err := e.FindAllByRegex(pattern, opts)
	if err != nil {
		return nil, false
	}
	return firstElement(selection)
}

func (e *Element) FindAllByRegex(pattern any, opts TextSearchOptions) (Selection, error) {
	if e == nil || e.node == nil {
		return Selection{}, nil
	}
	return findAllByRegex(e.doc, e.node, pattern, opts)
}

func (e *Element) FindSimilar(opts SimilarOptions) Selection {
	if e == nil || e.doc == nil || e.doc.root == nil || e.node == nil || e.node.Type != html.ElementNode {
		return Selection{}
	}
	threshold := opts.Threshold
	if threshold == 0 {
		threshold = 0.2
	}
	ignoreAttributes := opts.IgnoreAttributes
	if ignoreAttributes == nil {
		ignoreAttributes = []string{"href", "src"}
	}

	parent := parentElementNode(e.node)
	grandparent := parentElementNode(parent)
	depth := elementDepth(e.node)
	targetAttrs := comparableAttrs(e.node, ignoreAttributes)
	targetText := directElementText(e.node, true)

	matches := make([]*Element, 0)
	for _, candidate := range allDescendantElements(e.doc, e.doc.root) {
		if candidate.node == e.node || candidate.node.Type != html.ElementNode {
			continue
		}
		if candidate.TagName() != e.TagName() || elementDepth(candidate.node) != depth {
			continue
		}
		if !sameNodeTag(parentElementNode(candidate.node), parent) ||
			!sameNodeTag(parentElementNode(parentElementNode(candidate.node)), grandparent) {
			continue
		}

		score := attributeSimilarity(targetAttrs, comparableAttrs(candidate.node, ignoreAttributes))
		if opts.MatchText {
			score = (score + stringSimilarity(targetText, directElementText(candidate.node, true))) / 2
		}
		if score >= threshold {
			matches = append(matches, candidate)
		}
	}
	return Selection{elements: matches}
}

func findAllByText(doc *Document, root *html.Node, text string, opts TextSearchOptions) Selection {
	query := normalizeSearchText(text, opts)
	matches := make([]*Element, 0)
	for _, element := range allDescendantElements(doc, root) {
		nodeText := normalizeSearchText(directElementText(element.node, opts.cleanMatch()), opts)
		if opts.Partial && strings.Contains(nodeText, query) || !opts.Partial && nodeText == query {
			matches = append(matches, element)
		}
	}
	return Selection{elements: matches}
}

func findAllByRegex(doc *Document, root *html.Node, pattern any, opts TextSearchOptions) (Selection, error) {
	expr, err := compileSearchRegex(pattern, opts)
	if err != nil {
		return Selection{}, err
	}
	matches := make([]*Element, 0)
	for _, element := range allDescendantElements(doc, root) {
		nodeText := directElementText(element.node, opts.cleanMatch())
		if expr.MatchString(nodeText) {
			matches = append(matches, element)
		}
	}
	return Selection{elements: matches}, nil
}

func firstElement(selection Selection) (*Element, bool) {
	return selection.First()
}

func (opts TextSearchOptions) cleanMatch() bool {
	if opts.CleanMatch == nil {
		return true
	}
	return *opts.CleanMatch
}

func normalizeSearchText(text string, opts TextSearchOptions) string {
	if opts.cleanMatch() {
		text = normalizeSpace(text)
	}
	if !opts.CaseSensitive {
		text = strings.ToLower(text)
	}
	return text
}

func compileSearchRegex(pattern any, opts TextSearchOptions) (*regexp.Regexp, error) {
	switch value := pattern.(type) {
	case string:
		if !opts.CaseSensitive {
			value = "(?i)" + value
		}
		return regexp.Compile(value)
	case *regexp.Regexp:
		return value, nil
	default:
		return nil, fmt.Errorf("%w: unsupported regex pattern type %T", ErrInvalidSelector, pattern)
	}
}

func directElementText(node *html.Node, clean bool) string {
	if node == nil {
		return ""
	}
	parts := make([]string, 0)
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.TextNode {
			continue
		}
		text := child.Data
		if clean {
			text = normalizeSpace(text)
		} else {
			text = strings.TrimSpace(text)
		}
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func comparableAttrs(node *html.Node, ignore []string) map[string]string {
	attrs := map[string]string{}
	if node == nil {
		return attrs
	}
	for _, attr := range node.Attr {
		key := strings.ToLower(attr.Key)
		if slices.Contains(ignore, key) {
			continue
		}
		attrs[key] = attr.Val
	}
	return attrs
}

func sameNodeTag(left *html.Node, right *html.Node) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Type == html.ElementNode && right.Type == html.ElementNode && strings.EqualFold(left.Data, right.Data)
}

func attributeSimilarity(left map[string]string, right map[string]string) float64 {
	return (mapKeySimilarity(left, right) + attrValueSimilarity(left, right)) / 2
}
