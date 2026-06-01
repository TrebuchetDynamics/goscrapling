package gormes

import (
	"github.com/TrebuchetDynamics/goscrapling/engines/browser"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes/browsertools"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes/contract"
)

const (
	ToolRenderedMarkdown    = browsertools.ToolRenderedMarkdown
	ToolLinks               = browsertools.ToolLinks
	ToolSemanticTree        = browsertools.ToolSemanticTree
	ToolStructuredData      = browsertools.ToolStructuredData
	ToolInteractiveElements = browsertools.ToolInteractiveElements
)

var ErrUnknownTool = contract.ErrUnknownTool

type BrowserExtractionAdapter = browsertools.BrowserExtractionAdapter

type ToolCall = browsertools.ToolCall

type ToolResult = browsertools.ToolResult

type Link = browser.Link

type StructuredData = browser.StructuredData
