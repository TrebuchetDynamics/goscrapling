package goscrapling

import (
	"strings"

	"golang.org/x/net/html"
)

type Fingerprint struct {
	Tag              string
	Text             string
	Attributes       map[string]string
	ParentTag        string
	ParentText       string
	ParentAttributes map[string]string
	SiblingTags      []string
	PathTags         []string
}

func fingerprintNode(node *html.Node) Fingerprint {
	if node == nil {
		return Fingerprint{}
	}

	fp := Fingerprint{
		Tag:              tagName(node),
		Text:             nodeText(node),
		Attributes:       nodeAttributes(node),
		ParentAttributes: map[string]string{},
		SiblingTags:      siblingTags(node),
		PathTags:         pathTags(node),
	}

	if parent := parentElement(node); parent != nil {
		fp.ParentTag = tagName(parent)
		fp.ParentText = nodeText(parent)
		fp.ParentAttributes = nodeAttributes(parent)
	}

	return fp
}

func tagName(node *html.Node) string {
	if node == nil || node.Type != html.ElementNode {
		return ""
	}
	return strings.ToLower(node.Data)
}

func nodeAttributes(node *html.Node) map[string]string {
	attrs := make(map[string]string)
	if node == nil {
		return attrs
	}

	for _, attr := range node.Attr {
		key := strings.ToLower(strings.TrimSpace(attr.Key))
		if key == "" {
			continue
		}
		attrs[key] = normalizeSpace(attr.Val)
	}

	return attrs
}

func parentElement(node *html.Node) *html.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Type == html.ElementNode {
			return parent
		}
	}
	return nil
}

func siblingTags(node *html.Node) []string {
	if node == nil || node.Parent == nil {
		return nil
	}

	var tags []string
	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling == node || sibling.Type != html.ElementNode {
			continue
		}
		tags = append(tags, tagName(sibling))
	}

	return tags
}

func pathTags(node *html.Node) []string {
	if node == nil {
		return nil
	}

	var reversed []string
	for current := node; current != nil; current = current.Parent {
		if current.Type == html.ElementNode {
			reversed = append(reversed, tagName(current))
		}
	}

	path := make([]string, 0, len(reversed))
	for i := len(reversed) - 1; i >= 0; i-- {
		path = append(path, reversed[i])
	}

	return path
}
