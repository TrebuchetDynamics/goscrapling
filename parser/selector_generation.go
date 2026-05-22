package parser

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// GenerateCSSSelector returns a deterministic CSS selector for the element.
// It follows Scrapling's selector-generation rule of stopping at the nearest
// id-bearing ancestor and using nth-of-type only when needed for same-tag
// siblings.
func (e *Element) GenerateCSSSelector() string {
	return e.generalSelector(selectorFormatCSS, false)
}

// GenerateFullCSSSelector returns a deterministic CSS selector from the body
// path to the element instead of stopping at the nearest id-bearing ancestor.
func (e *Element) GenerateFullCSSSelector() string {
	return e.generalSelector(selectorFormatCSS, true)
}

// GenerateXPathSelector returns a deterministic XPath selector for the element,
// stopping at the nearest id-bearing ancestor when one exists.
func (e *Element) GenerateXPathSelector() string {
	return e.generalSelector(selectorFormatXPath, false)
}

// GenerateFullXPathSelector returns a deterministic XPath selector from the body
// path to the element instead of stopping at the nearest id-bearing ancestor.
func (e *Element) GenerateFullXPathSelector() string {
	return e.generalSelector(selectorFormatXPath, true)
}

type selectorFormat int

const (
	selectorFormatCSS selectorFormat = iota
	selectorFormatXPath
)

func (e *Element) generalSelector(format selectorFormat, fullPath bool) string {
	if e == nil || e.node == nil || e.node.Type != html.ElementNode {
		return ""
	}

	parts := make([]string, 0)
	for target := e.node; target != nil && target.Type == html.ElementNode; target = parentElementNode(target) {
		parent := parentElementNode(target)
		if parent == nil {
			break
		}

		if id, ok := nodeAttr(target, "id"); ok && id != "" {
			parts = append(parts, selectorIDPart(format, fullPath, id))
			if !fullPath {
				return joinSelectorParts(format, parts, true)
			}
		} else {
			parts = append(parts, selectorTagPart(format, target))
		}

		if nodeTagName(parent) == "html" {
			return joinSelectorParts(format, parts, false)
		}
	}

	return joinSelectorParts(format, parts, false)
}

func selectorIDPart(format selectorFormat, fullPath bool, id string) string {
	if format == selectorFormatCSS {
		return "#" + cssIdentifier(id)
	}
	if fullPath {
		return "*[@id=" + xpathLiteral(id) + "]"
	}
	return "[@id=" + xpathLiteral(id) + "]"
}

func selectorTagPart(format selectorFormat, node *html.Node) string {
	part := nodeTagName(node)
	index := nthOfTypeIndex(node)
	if index <= 1 {
		return part
	}
	if format == selectorFormatCSS {
		return fmt.Sprintf("%s:nth-of-type(%d)", part, index)
	}
	return fmt.Sprintf("%s[%d]", part, index)
}

func joinSelectorParts(format selectorFormat, parts []string, startsAtID bool) string {
	if len(parts) == 0 {
		return ""
	}
	reversed := make([]string, len(parts))
	for i := range parts {
		reversed[i] = parts[len(parts)-1-i]
	}
	if format == selectorFormatCSS {
		return strings.Join(reversed, " > ")
	}
	if startsAtID {
		return "//*" + strings.Join(reversed, "/")
	}
	return "//" + strings.Join(reversed, "/")
}

func nodeAttr(node *html.Node, name string) (string, bool) {
	if node == nil {
		return "", false
	}
	name = strings.ToLower(name)
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == name {
			return attr.Val, true
		}
	}
	return "", false
}

func nodeTagName(node *html.Node) string {
	if node == nil || node.Type != html.ElementNode {
		return ""
	}
	return strings.ToLower(node.Data)
}

func nthOfTypeIndex(node *html.Node) int {
	if node == nil || node.Parent == nil || node.Type != html.ElementNode {
		return 0
	}
	index := 0
	name := nodeTagName(node)
	for sibling := node.Parent.FirstChild; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type != html.ElementNode || nodeTagName(sibling) != name {
			continue
		}
		index++
		if sibling == node {
			return index
		}
	}
	return index
}

func cssIdentifier(value string) string {
	var b strings.Builder
	for i, r := range value {
		if isSafeCSSIdentifierRune(r) && !(i == 0 && unicode.IsDigit(r)) {
			b.WriteRune(r)
			continue
		}
		fmt.Fprintf(&b, `\%X `, r)
	}
	return b.String()
}

func isSafeCSSIdentifierRune(r rune) bool {
	return r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func xpathLiteral(value string) string {
	if !strings.Contains(value, `'`) {
		return `'` + value + `'`
	}
	if !strings.Contains(value, `"`) {
		return `"` + value + `"`
	}
	parts := strings.Split(value, `'`)
	quoted := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if part != "" {
			quoted = append(quoted, `'`+part+`'`)
		}
		if i != len(parts)-1 {
			quoted = append(quoted, `"'"`)
		}
	}
	return "concat(" + strings.Join(quoted, ", ") + ")"
}
