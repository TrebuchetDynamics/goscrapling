package browsertools

import (
	"context"
	"net/http"

	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes/contract"
)

const (
	ToolRenderedMarkdown    = "rendered_markdown"
	ToolLinks               = "links"
	ToolSemanticTree        = "semantic_tree"
	ToolStructuredData      = "structured_data"
	ToolInteractiveElements = "interactive_elements"
)

var ErrUnknownTool = contract.ErrUnknownTool

type BrowserExtractionAdapter struct {
	Engine browser.BrowserEngine
}

type ToolCall struct {
	Tool string
	URL  string
	Opts browser.BrowserOptions
}

type ToolResult struct {
	Tool                string
	URL                 string
	StatusCode          int
	Markdown            string
	Links               []Link
	SemanticTree        []browser.SemanticNode
	StructuredData      StructuredData
	InteractiveElements []browser.SemanticNode
}

type Link = browser.Link

type StructuredData = browser.StructuredData

func (a BrowserExtractionAdapter) Call(ctx context.Context, call ToolCall) (ToolResult, error) {
	fetcher := browser.BrowserFetcher{Engine: a.Engine}
	if call.Opts.Headers == nil {
		call.Opts.Headers = http.Header{}
	}

	switch call.Tool {
	case ToolRenderedMarkdown:
		dump, err := fetcher.FetchMarkdown(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{Tool: call.Tool, URL: dump.URL, StatusCode: dump.StatusCode, Markdown: dump.Markdown}, nil
	case ToolSemanticTree, ToolInteractiveElements:
		tree, err := fetcher.FetchSemanticTree(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		result := ToolResult{Tool: call.Tool, URL: tree.URL, StatusCode: tree.StatusCode, SemanticTree: tree.Nodes}
		if call.Tool == ToolInteractiveElements {
			result.SemanticTree = nil
			for _, node := range tree.Nodes {
				if node.Interactive {
					result.InteractiveElements = append(result.InteractiveElements, node)
				}
			}
		}
		return result, nil
	case ToolLinks, ToolStructuredData:
		response, err := fetcher.Fetch(ctx, call.URL, call.Opts)
		if err != nil {
			return ToolResult{}, err
		}
		evidence, err := browser.NewRenderedEvidence(response.Body())
		if err != nil {
			return ToolResult{}, err
		}
		result := ToolResult{Tool: call.Tool, URL: response.URL(), StatusCode: response.StatusCode()}
		if call.Tool == ToolLinks {
			result.Links = evidence.Links()
		} else {
			result.StructuredData = evidence.StructuredData()
		}
		return result, nil
	default:
		return ToolResult{}, ErrUnknownTool
	}
}
