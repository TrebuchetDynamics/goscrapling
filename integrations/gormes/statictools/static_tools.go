package statictools

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/TrebuchetDynamics/goscrapling/integrations/gormes/contract"
)

const ToolWebExtract = "web_extract"

var ErrUnknownTool = contract.ErrUnknownTool

type StaticExtractionAdapter struct {
	Fetcher   fetchers.Fetcher
	Recipes   map[string]ExtractionRecipe
	Providers []StaticProvider
}

type StaticToolCall struct {
	Tool         string
	URLs         []string
	CSSSelector  string
	Provider     string
	Recipe       string
	RecipeParams map[string]string
	Pageno       int
	Language     string
	SafeSearch   int
	TimeRange    string
	Opts         fetchers.RequestOptions
}

type WebExtractResponse struct {
	Results []WebExtractResult `json:"results"`
}

type WebExtractResult struct {
	URL         string                    `json:"url"`
	Title       string                    `json:"title"`
	Content     string                    `json:"content"`
	Fields      map[string]ExtractedField `json:"fields,omitempty"`
	Suggestions []string                  `json:"suggestions,omitempty"`
	Error       string                    `json:"error,omitempty"`
	Extraction  *WebExtraction            `json:"extraction,omitempty"`
}

type WebExtraction struct {
	Engine           string        `json:"engine,omitempty"`
	Mode             string        `json:"mode,omitempty"`
	Recipe           string        `json:"recipe,omitempty"`
	Provider         string        `json:"provider,omitempty"`
	ProviderShortcut string        `json:"provider_shortcut,omitempty"`
	ProviderWeight   float64       `json:"provider_weight,omitempty"`
	ProviderTimeout  time.Duration `json:"provider_timeout,omitempty"`
	StatusCode       int           `json:"status_code,omitempty"`
	ContentType      string        `json:"content_type,omitempty"`
	CSSSelector      string        `json:"css_selector,omitempty"`
	FinalURL         string        `json:"final_url,omitempty"`
	NoResult         bool          `json:"no_result,omitempty"`
}

func (a StaticExtractionAdapter) Call(ctx context.Context, call StaticToolCall) (WebExtractResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if call.Tool != ToolWebExtract {
		return WebExtractResponse{}, ErrUnknownTool
	}
	if strings.TrimSpace(call.Provider) != "" {
		if err := ctx.Err(); err != nil {
			return WebExtractResponse{}, err
		}
		result, err := a.extractProvider(ctx, call)
		if err != nil {
			return WebExtractResponse{}, err
		}
		return WebExtractResponse{Results: []WebExtractResult{result}}, nil
	}
	if strings.TrimSpace(call.Recipe) != "" {
		if err := ctx.Err(); err != nil {
			return WebExtractResponse{}, err
		}
		result, err := a.extractRecipe(ctx, call)
		if err != nil {
			return WebExtractResponse{}, err
		}
		return WebExtractResponse{Results: []WebExtractResult{result}}, nil
	}
	if len(call.URLs) == 0 {
		return WebExtractResponse{}, fmt.Errorf("%s: urls is required", ToolWebExtract)
	}

	results := make([]WebExtractResult, 0, len(call.URLs))
	for _, rawURL := range call.URLs {
		if err := ctx.Err(); err != nil {
			return WebExtractResponse{}, err
		}
		trimmedURL := strings.TrimSpace(rawURL)
		if trimmedURL == "" {
			results = append(results, WebExtractResult{Error: "url is required"})
			continue
		}
		result, err := a.extractURL(trimmedURL, call)
		if err != nil {
			results = append(results, WebExtractResult{URL: trimmedURL, Error: err.Error()})
			continue
		}
		results = append(results, result)
	}
	return WebExtractResponse{Results: results}, nil
}

func (a StaticExtractionAdapter) extractURL(rawURL string, call StaticToolCall) (WebExtractResult, error) {
	response, err := a.Fetcher.Get(rawURL, call.Opts)
	if err != nil {
		return WebExtractResult{}, err
	}

	finalURL := response.URL()
	if finalURL == "" {
		finalURL = rawURL
	}
	selector := strings.TrimSpace(call.CSSSelector)
	extraction := &WebExtraction{
		Engine:      "goscrapling",
		Mode:        "static",
		StatusCode:  response.StatusCode(),
		ContentType: mediaType(response.Headers().Get("Content-Type")),
		CSSSelector: selector,
		FinalURL:    finalURL,
	}
	result := WebExtractResult{
		URL:        finalURL,
		Title:      strings.TrimSpace(response.CSS("title::text").Get().String()),
		Extraction: extraction,
	}
	if status := response.StatusCode(); status < http.StatusOK || status >= http.StatusMultipleChoices {
		result.Error = fmt.Sprintf("HTTP %d", status)
		return result, nil
	}

	selection := response.CSS("body")
	if selector != "" {
		selection = response.CSS(selector)
		if selection.Len() == 0 {
			result.Error = "css_selector matched no elements"
			return result, nil
		}
	}
	content := strings.TrimSpace(selection.Text())
	if content == "" {
		if html, err := selection.HTML(); err == nil {
			content = strings.TrimSpace(html)
		}
	}
	if content == "" {
		result.Error = "empty extracted content"
		return result, nil
	}
	result.Content = content
	return result, nil
}

func mediaType(contentType string) string {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(contentType))
	}
	return strings.ToLower(strings.TrimSpace(mediaType))
}
