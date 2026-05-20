package browser

import (
	"bytes"
	"context"
	"strings"

	"golang.org/x/net/html"
)

type BrowserMarkdownDump struct {
	URL        string
	StatusCode int
	Markdown   string
	Response   *Response
}

func (f BrowserFetcher) FetchMarkdown(ctx context.Context, rawURL string, opts BrowserOptions) (BrowserMarkdownDump, error) {
	response, err := f.Fetch(ctx, rawURL, opts)
	if err != nil {
		return BrowserMarkdownDump{}, err
	}
	markdown, err := HTMLToMarkdown(response.Body())
	if err != nil {
		return BrowserMarkdownDump{}, err
	}
	return BrowserMarkdownDump{
		URL:        response.URL(),
		StatusCode: response.StatusCode(),
		Markdown:   markdown,
		Response:   response,
	}, nil
}

func HTMLToMarkdown(body []byte) (string, error) {
	root, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	var renderer markdownRenderer
	renderer.renderBlocks(root)
	return strings.TrimSpace(renderer.output.String()), nil
}

type markdownRenderer struct {
	output strings.Builder
}

func (r *markdownRenderer) renderBlocks(node *html.Node) {
	if node == nil || skipMarkdownNode(node) {
		return
	}
	if node.Type == html.ElementNode {
		switch strings.ToLower(node.Data) {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			level := int(node.Data[1] - '0')
			r.writeBlock(strings.Repeat("#", level) + " " + inlineText(node))
			return
		case "p":
			r.writeBlock(inlineText(node))
			return
		case "li":
			r.writeBlock("- " + inlineText(node))
			return
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		r.renderBlocks(child)
	}
}

func (r *markdownRenderer) writeBlock(value string) {
	value = normalizeMarkdownSpace(value)
	if value == "" {
		return
	}
	if r.output.Len() > 0 {
		r.output.WriteString("\n\n")
	}
	r.output.WriteString(value)
}

func inlineText(node *html.Node) string {
	var builder strings.Builder
	writeInline(&builder, node)
	return normalizeMarkdownSpace(builder.String())
}

func writeInline(builder *strings.Builder, node *html.Node) {
	if node == nil || skipMarkdownNode(node) {
		return
	}
	if node.Type == html.TextNode {
		builder.WriteByte(' ')
		builder.WriteString(node.Data)
		builder.WriteByte(' ')
		return
	}
	if node.Type == html.ElementNode && strings.EqualFold(node.Data, "a") {
		label := inlineChildrenText(node)
		href := attrValue(node, "href")
		if label != "" && href != "" {
			builder.WriteString(" [")
			builder.WriteString(label)
			builder.WriteString("](")
			builder.WriteString(href)
			builder.WriteString(") ")
			return
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeInline(builder, child)
	}
}

func inlineChildrenText(node *html.Node) string {
	var builder strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		writeInline(&builder, child)
	}
	return normalizeMarkdownSpace(builder.String())
}

func skipMarkdownNode(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	switch strings.ToLower(node.Data) {
	case "script", "style", "noscript", "template", "svg", "nav", "header", "footer", "aside":
		return true
	default:
		return false
	}
}

func attrValue(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if strings.EqualFold(attr.Key, name) {
			return attr.Val
		}
	}
	return ""
}

func normalizeMarkdownSpace(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	value = strings.NewReplacer(
		" .", ".",
		" ,", ",",
		" !", "!",
		" ?", "?",
		" ;", ";",
		" :", ":",
	).Replace(value)
	return value
}
