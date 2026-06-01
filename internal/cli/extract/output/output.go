package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling"
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
	"golang.org/x/net/html"
)

func Render(response *goscrapling.Response, outputPath string, selector string) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".html", ".htm":
		if selector == "" {
			return response.Body(), nil
		}
		selectedHTML, err := response.CSS(selector).HTML()
		if err != nil {
			return nil, err
		}
		return []byte(selectedHTML), nil
	case ".md":
		htmlBody := response.Body()
		if selector != "" {
			selectedHTML, err := response.CSS(selector).HTML()
			if err != nil {
				return nil, err
			}
			htmlBody = []byte(selectedHTML)
		}
		markdown, err := browser.HTMLToMarkdown(htmlBody)
		if err != nil {
			return nil, err
		}
		return []byte(markdown), nil
	case ".txt", "":
		return []byte(extractText(response, selector)), nil
	default:
		return nil, diagnostics.ParseError("unsupported output extension %q", ext)
	}
}

func extractText(response *goscrapling.Response, selector string) string {
	if selector != "" {
		return response.CSS(selector).Text()
	}
	bodyText := response.CSS("body").Text()
	if bodyText != "" {
		return bodyText
	}
	return strings.TrimSpace(string(bytes.TrimSpace(response.Body())))
}

func AITargetedResponse(response *goscrapling.Response) (*goscrapling.Response, error) {
	cleaned, err := aiTargetedHTML(response.Body())
	if err != nil {
		return nil, err
	}
	return goscrapling.NewResponse(bytes.NewReader(cleaned), goscrapling.ResponseOptions{
		URL:         response.URL(),
		StatusCode:  response.StatusCode(),
		Reason:      response.Reason(),
		Headers:     response.Headers(),
		Request:     response.Request(),
		Encoding:    response.Encoding(),
		Cookies:     response.Cookies(),
		History:     response.History(),
		Meta:        response.Meta(),
		CapturedXHR: response.CapturedXHR(),
	})
}

func aiTargetedHTML(body []byte) ([]byte, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	stripAITargetedNoise(root)
	target := firstElementByTag(root, "main")
	if target == nil {
		target = firstElementByTag(root, "body")
	}
	if target == nil {
		target = root
	}
	var cleaned bytes.Buffer
	if err := html.Render(&cleaned, target); err != nil {
		return nil, err
	}
	return cleaned.Bytes(), nil
}

func stripAITargetedNoise(node *html.Node) {
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		if shouldRemoveAITargetedNode(child) {
			node.RemoveChild(child)
		} else {
			if child.Type == html.TextNode {
				child.Data = removeZeroWidthRunes(child.Data)
			}
			stripAITargetedNoise(child)
		}
		child = next
	}
}

func shouldRemoveAITargetedNode(node *html.Node) bool {
	if node.Type == html.CommentNode {
		return true
	}
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "script", "style", "noscript", "svg", "template", "nav", "header", "footer", "aside":
		return true
	}
	if _, ok := htmlAttr(node, "hidden"); ok {
		return true
	}
	if value, ok := htmlAttr(node, "aria-hidden"); ok && strings.EqualFold(strings.TrimSpace(value), "true") {
		return true
	}
	if strings.EqualFold(node.Data, "input") {
		if value, ok := htmlAttr(node, "type"); ok && strings.EqualFold(strings.TrimSpace(value), "hidden") {
			return true
		}
	}
	if style, ok := htmlAttr(node, "style"); ok {
		normalized := strings.ToLower(strings.ReplaceAll(style, " ", ""))
		if strings.Contains(normalized, "display:none") || strings.Contains(normalized, "visibility:hidden") {
			return true
		}
	}
	return false
}

func firstElementByTag(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, tag) {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := firstElementByTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func htmlAttr(node *html.Node, name string) (string, bool) {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val, true
		}
	}
	return "", false
}

func removeZeroWidthRunes(value string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u200b', '\u200c', '\u200d', '\ufeff', '\u2060':
			return -1
		default:
			return r
		}
	}, value)
}

func WriteFile(path string, body []byte) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, body, 0o644)
}
