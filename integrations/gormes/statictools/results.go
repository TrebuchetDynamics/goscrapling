package statictools

import (
	"sort"
	"strings"
)

type ResultContainer struct {
	results       map[string]*MergedResult
	order         []string
	position      int
	suggestions   []string
	suggestionSet map[string]struct{}
	unresponsive  []UnresponsiveProvider
}

type ResultContainerSnapshot struct {
	Results               []MergedResult         `json:"results"`
	Suggestions           []string               `json:"suggestions,omitempty"`
	UnresponsiveProviders []UnresponsiveProvider `json:"unresponsive_providers,omitempty"`
}

type MergedResult struct {
	URL         string                    `json:"url"`
	Title       string                    `json:"title"`
	Content     string                    `json:"content"`
	Fields      map[string]ExtractedField `json:"fields,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
	Providers   []ResultProviderEvidence  `json:"providers,omitempty"`
	Score       float64                   `json:"score"`
}

type ResultProviderEvidence struct {
	Provider string                    `json:"provider"`
	Position int                       `json:"position"`
	Weight   float64                   `json:"weight"`
	Fields   map[string]ExtractedField `json:"fields,omitempty"`
}

type UnresponsiveProvider struct {
	Provider string `json:"provider"`
	Error    string `json:"error"`
}

func NewResultContainer() *ResultContainer {
	return &ResultContainer{
		results:       map[string]*MergedResult{},
		suggestionSet: map[string]struct{}{},
	}
}

func (c *ResultContainer) Add(result WebExtractResult) {
	if c == nil {
		return
	}
	if c.results == nil {
		c.results = map[string]*MergedResult{}
	}
	if c.suggestionSet == nil {
		c.suggestionSet = map[string]struct{}{}
	}

	c.position++
	key := resultContainerKey(result)
	if key == "" {
		key = "#" + resultProviderName(result) + ":" + strings.TrimSpace(result.Title)
	}
	merged, exists := c.results[key]
	if !exists {
		merged = &MergedResult{
			URL:     resultContainerURL(result),
			Title:   strings.TrimSpace(result.Title),
			Content: strings.TrimSpace(result.Content),
			Fields:  cloneExtractedFields(result.Fields),
		}
		c.results[key] = merged
		c.order = append(c.order, key)
	}
	if merged.URL == "" {
		merged.URL = resultContainerURL(result)
	}
	if merged.Title == "" {
		merged.Title = strings.TrimSpace(result.Title)
	}
	if merged.Content == "" {
		merged.Content = strings.TrimSpace(result.Content)
	}
	mergeMissingFields(merged.Fields, result.Fields)

	provider := ResultProviderEvidence{
		Provider: resultProviderName(result),
		Position: c.position,
		Weight:   resultProviderWeight(result),
		Fields:   cloneExtractedFields(result.Fields),
	}
	merged.Providers = append(merged.Providers, provider)
	merged.Score += provider.Weight / float64(provider.Position)

	for _, suggestion := range result.Suggestions {
		suggestion = strings.TrimSpace(suggestion)
		if suggestion == "" {
			continue
		}
		if !stringInSlice(merged.Suggestions, suggestion) {
			merged.Suggestions = append(merged.Suggestions, suggestion)
		}
		if _, exists := c.suggestionSet[suggestion]; !exists {
			c.suggestionSet[suggestion] = struct{}{}
			c.suggestions = append(c.suggestions, suggestion)
		}
	}
}

func (c *ResultContainer) AddUnresponsiveProvider(provider string, err string) {
	if c == nil {
		return
	}
	provider = strings.TrimSpace(provider)
	err = strings.TrimSpace(err)
	if provider == "" && err == "" {
		return
	}
	c.unresponsive = append(c.unresponsive, UnresponsiveProvider{Provider: provider, Error: err})
}

func (c *ResultContainer) Snapshot() ResultContainerSnapshot {
	if c == nil {
		return ResultContainerSnapshot{}
	}
	results := make([]MergedResult, 0, len(c.results))
	for _, key := range c.order {
		merged, ok := c.results[key]
		if !ok || merged == nil {
			continue
		}
		results = append(results, cloneMergedResult(*merged))
	}
	sort.SliceStable(results, func(left, right int) bool {
		if results[left].Score == results[right].Score {
			return firstProviderPosition(results[left]) < firstProviderPosition(results[right])
		}
		return results[left].Score > results[right].Score
	})
	return ResultContainerSnapshot{
		Results:               results,
		Suggestions:           append([]string(nil), c.suggestions...),
		UnresponsiveProviders: append([]UnresponsiveProvider(nil), c.unresponsive...),
	}
}

func (r MergedResult) ProviderNames() []string {
	names := make([]string, 0, len(r.Providers))
	for _, provider := range r.Providers {
		names = append(names, provider.Provider)
	}
	return names
}

func resultContainerKey(result WebExtractResult) string {
	return strings.TrimSpace(resultContainerURL(result))
}

func resultContainerURL(result WebExtractResult) string {
	if result.Extraction != nil && strings.TrimSpace(result.Extraction.FinalURL) != "" {
		return strings.TrimSpace(result.Extraction.FinalURL)
	}
	return strings.TrimSpace(result.URL)
}

func resultProviderName(result WebExtractResult) string {
	if result.Extraction != nil && strings.TrimSpace(result.Extraction.Provider) != "" {
		return strings.TrimSpace(result.Extraction.Provider)
	}
	return "unknown"
}

func resultProviderWeight(result WebExtractResult) float64 {
	if result.Extraction != nil && result.Extraction.ProviderWeight > 0 {
		return result.Extraction.ProviderWeight
	}
	return 1
}

func mergeMissingFields(target map[string]ExtractedField, source map[string]ExtractedField) {
	if len(source) == 0 {
		return
	}
	if target == nil {
		return
	}
	for key, value := range source {
		if _, exists := target[key]; !exists {
			target[key] = value
		}
	}
}

func cloneExtractedFields(fields map[string]ExtractedField) map[string]ExtractedField {
	if len(fields) == 0 {
		return nil
	}
	clone := make(map[string]ExtractedField, len(fields))
	for key, value := range fields {
		value.Values = append([]string(nil), value.Values...)
		clone[key] = value
	}
	return clone
}

func cloneMergedResult(result MergedResult) MergedResult {
	result.Fields = cloneExtractedFields(result.Fields)
	result.Suggestions = append([]string(nil), result.Suggestions...)
	result.Providers = append([]ResultProviderEvidence(nil), result.Providers...)
	for index := range result.Providers {
		result.Providers[index].Fields = cloneExtractedFields(result.Providers[index].Fields)
	}
	return result
}

func firstProviderPosition(result MergedResult) int {
	if len(result.Providers) == 0 {
		return 0
	}
	return result.Providers[0].Position
}

func stringInSlice(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
