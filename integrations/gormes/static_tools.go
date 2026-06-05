package gormes

import "github.com/TrebuchetDynamics/goscrapling/integrations/gormes/statictools"

const ToolWebExtract = statictools.ToolWebExtract

type StaticExtractionAdapter = statictools.StaticExtractionAdapter

type StaticToolCall = statictools.StaticToolCall

type WebExtractResponse = statictools.WebExtractResponse

type WebExtractResult = statictools.WebExtractResult

type WebExtraction = statictools.WebExtraction

type SelectorType = statictools.SelectorType

const (
	SelectorCSS   = statictools.SelectorCSS
	SelectorXPath = statictools.SelectorXPath
	SelectorJSON  = statictools.SelectorJSON
)

type ExtractionRecipe = statictools.ExtractionRecipe

type ExtractionField = statictools.ExtractionField

type ExtractedField = statictools.ExtractedField
