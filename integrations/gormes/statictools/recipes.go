package statictools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/fetchers"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type SelectorType string

const (
	SelectorCSS   SelectorType = "css"
	SelectorXPath SelectorType = "xpath"
	SelectorJSON  SelectorType = "json"
)

type ExtractionRecipe struct {
	URLTemplate         string
	Params              url.Values
	Method              string
	Headers             http.Header
	Cookies             map[string]string
	BodyTemplate        string
	PageSize            int
	FirstPageNum        int
	LangAll             string
	SafeSearchMap       map[int]string
	TimeRangeTemplate   string
	TimeRangeMap        map[string]string
	NoResultStatusCodes []int
	Fields              []ExtractionField
	Suggestions         []ExtractionField
}

type ExtractionField struct {
	Name       string
	Type       SelectorType
	Selector   string
	Multiple   bool
	Required   bool
	Prefix     string
	HTMLToText bool
}

type ExtractedField struct {
	Type     SelectorType `json:"type,omitempty"`
	Selector string       `json:"selector,omitempty"`
	Value    string       `json:"value,omitempty"`
	Values   []string     `json:"values,omitempty"`
	Error    string       `json:"error,omitempty"`
}

var recipePlaceholderPattern = regexp.MustCompile(`\{([A-Za-z0-9_.-]+)\}`)

func (a StaticExtractionAdapter) extractRecipe(ctx context.Context, call StaticToolCall) (WebExtractResult, error) {
	recipeName := strings.TrimSpace(call.Recipe)
	recipe, ok := a.Recipes[recipeName]
	if !ok {
		return WebExtractResult{}, fmt.Errorf("%s: unknown recipe %q", ToolWebExtract, recipeName)
	}

	values, err := recipeTemplateValues(recipe, call)
	if err != nil {
		return WebExtractResult{}, err
	}
	rawURL, err := recipeURL(recipe, values)
	if err != nil {
		return WebExtractResult{}, err
	}
	opts, err := recipeRequestOptions(recipe, call.Opts, values)
	if err != nil {
		return WebExtractResult{}, err
	}
	if opts.Context == nil {
		opts.Context = ctx
	}

	response, err := a.fetchRecipe(rawURL, recipe, opts)
	if err != nil {
		return WebExtractResult{}, err
	}

	finalURL := response.URL()
	if finalURL == "" {
		finalURL = rawURL
	}
	extraction := &WebExtraction{
		Engine:      "goscrapling",
		Mode:        "static",
		Recipe:      recipeName,
		StatusCode:  response.StatusCode(),
		ContentType: mediaType(response.Headers().Get("Content-Type")),
		FinalURL:    finalURL,
	}
	result := WebExtractResult{
		URL:         finalURL,
		Title:       strings.TrimSpace(response.CSS("title::text").Get().String()),
		Extraction:  extraction,
		Suggestions: extractRecipeSuggestions(response, recipe),
	}
	if recipeNoResultForStatus(recipe, response.StatusCode()) {
		extraction.NoResult = true
		return result, nil
	}
	result.Fields = extractRecipeFields(response, recipe)
	if status := response.StatusCode(); status < 200 || status >= 300 {
		result.Error = fmt.Sprintf("HTTP %d", status)
	}
	return result, nil
}

func (a StaticExtractionAdapter) fetchRecipe(rawURL string, recipe ExtractionRecipe, opts fetchers.RequestOptions) (*fetchers.Response, error) {
	method := strings.ToUpper(strings.TrimSpace(recipe.Method))
	if method == "" {
		method = http.MethodGet
	}
	switch method {
	case http.MethodGet:
		return a.Fetcher.Get(rawURL, opts)
	case http.MethodPost:
		return a.Fetcher.Post(rawURL, opts)
	default:
		return nil, fmt.Errorf("%s: unsupported recipe method %q", ToolWebExtract, recipe.Method)
	}
}

func recipeURL(recipe ExtractionRecipe, values map[string]string) (string, error) {
	template := strings.TrimSpace(recipe.URLTemplate)
	if template == "" {
		return "", fmt.Errorf("%s: recipe url template is required", ToolWebExtract)
	}
	rendered, err := renderRecipeTemplate(template, values, recipeURLTemplateEscape)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(rendered)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	keys := make([]string, 0, len(recipe.Params))
	for key := range recipe.Params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		renderedKey, err := renderRecipeTemplate(key, values, noEscape)
		if err != nil {
			return "", err
		}
		for _, value := range recipe.Params[key] {
			renderedValue, err := renderRecipeTemplate(value, values, noEscape)
			if err != nil {
				return "", err
			}
			query.Add(renderedKey, renderedValue)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func recipeTemplateValues(recipe ExtractionRecipe, call StaticToolCall) (map[string]string, error) {
	values := make(map[string]string, len(call.RecipeParams)+5)
	for key, value := range call.RecipeParams {
		values[key] = value
	}
	values["pageno"] = strconv.Itoa(recipePageNumber(recipe, call))
	values["lang"] = recipeLanguage(recipe, call)
	values["safe_search"] = recipeSafeSearch(recipe, call)
	timeRange, timeRangeValue, err := recipeTimeRange(recipe, call)
	if err != nil {
		return nil, err
	}
	values["time_range"] = timeRange
	values["time_range_val"] = timeRangeValue
	return values, nil
}

func recipePageNumber(recipe ExtractionRecipe, call StaticToolCall) int {
	pageno := call.Pageno
	if pageno <= 0 {
		if raw := strings.TrimSpace(call.RecipeParams["pageno"]); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil {
				pageno = parsed
			}
		}
	}
	if pageno <= 0 {
		pageno = 1
	}
	pageSize := recipe.PageSize
	if pageSize <= 0 {
		pageSize = 1
	}
	firstPage := recipe.FirstPageNum
	if firstPage == 0 {
		firstPage = 1
	}
	return (pageno-1)*pageSize + firstPage
}

func recipeLanguage(recipe ExtractionRecipe, call StaticToolCall) string {
	language := strings.TrimSpace(call.Language)
	if language == "" {
		language = strings.TrimSpace(call.RecipeParams["lang"])
	}
	if language == "" || strings.EqualFold(language, "all") {
		if recipe.LangAll != "" {
			return recipe.LangAll
		}
		return "en"
	}
	language = strings.ToLower(strings.ReplaceAll(language, "_", "-"))
	if index := strings.Index(language, "-"); index > 0 {
		return language[:index]
	}
	return language
}

func recipeSafeSearch(recipe ExtractionRecipe, call StaticToolCall) string {
	level := call.SafeSearch
	if raw := strings.TrimSpace(call.RecipeParams["safesearch"]); raw != "" && call.SafeSearch == 0 {
		if parsed, err := strconv.Atoi(raw); err == nil {
			level = parsed
		}
	}
	if recipe.SafeSearchMap == nil {
		return ""
	}
	return recipe.SafeSearchMap[level]
}

func recipeTimeRange(recipe ExtractionRecipe, call StaticToolCall) (string, string, error) {
	rangeName := strings.TrimSpace(call.TimeRange)
	if rangeName == "" {
		rangeName = strings.TrimSpace(call.RecipeParams["time_range"])
	}
	if rangeName == "" || recipe.TimeRangeTemplate == "" {
		return "", "", nil
	}
	rangeValue := recipe.TimeRangeMap[rangeName]
	if rangeValue == "" {
		rangeValue = rangeName
	}
	values := map[string]string{"time_range_val": rangeValue}
	rendered, err := renderRecipeTemplate(recipe.TimeRangeTemplate, values, noEscape)
	if err != nil {
		return "", "", err
	}
	return rendered, rangeValue, nil
}

func recipeRequestOptions(recipe ExtractionRecipe, base fetchers.RequestOptions, values map[string]string) (fetchers.RequestOptions, error) {
	opts := base
	headers := opts.Headers.Clone()
	if headers == nil {
		headers = http.Header{}
	}
	for key, headerValues := range recipe.Headers {
		renderedKey, err := renderRecipeTemplate(key, values, noEscape)
		if err != nil {
			return opts, err
		}
		for _, value := range headerValues {
			renderedValue, err := renderRecipeTemplate(value, values, noEscape)
			if err != nil {
				return opts, err
			}
			headers.Add(renderedKey, renderedValue)
		}
	}
	opts.Headers = headers

	if len(recipe.Cookies) > 0 {
		cookies := make(map[string]string, len(opts.CookieValues)+len(recipe.Cookies))
		for name, value := range opts.CookieValues {
			cookies[name] = value
		}
		keys := make([]string, 0, len(recipe.Cookies))
		for key := range recipe.Cookies {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			renderedKey, err := renderRecipeTemplate(key, values, noEscape)
			if err != nil {
				return opts, err
			}
			renderedValue, err := renderRecipeTemplate(recipe.Cookies[key], values, noEscape)
			if err != nil {
				return opts, err
			}
			cookies[renderedKey] = renderedValue
		}
		opts.CookieValues = cookies
	}

	if strings.TrimSpace(recipe.BodyTemplate) != "" {
		body, err := renderRecipeTemplate(recipe.BodyTemplate, values, noEscape)
		if err != nil {
			return opts, err
		}
		opts.Body = strings.NewReader(body)
	}
	return opts, nil
}

func renderRecipeTemplate(template string, values map[string]string, escape func(string) string) (string, error) {
	missing := map[string]struct{}{}
	rendered := recipePlaceholderPattern.ReplaceAllStringFunc(template, func(token string) string {
		key := token[1 : len(token)-1]
		value, ok := values[key]
		if !ok {
			missing[key] = struct{}{}
			return token
		}
		return escape(value)
	})
	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for key := range missing {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return "", fmt.Errorf("%s: missing recipe params: %s", ToolWebExtract, strings.Join(keys, ", "))
	}
	return rendered, nil
}

func noEscape(value string) string { return value }

func recipeURLTemplateEscape(value string) string {
	if strings.HasPrefix(value, "&") || strings.HasPrefix(value, "?") {
		return value
	}
	return url.PathEscape(value)
}

func recipeNoResultForStatus(recipe ExtractionRecipe, status int) bool {
	for _, noResultStatus := range recipe.NoResultStatusCodes {
		if status == noResultStatus {
			return true
		}
	}
	return false
}

func extractRecipeSuggestions(response *fetchers.Response, recipe ExtractionRecipe) []string {
	if len(recipe.Suggestions) == 0 {
		return nil
	}
	var suggestions []string
	for _, field := range recipe.Suggestions {
		values, err := extractRecipeField(response, field)
		if err != nil {
			continue
		}
		suggestions = append(suggestions, values...)
	}
	if len(suggestions) == 0 {
		return nil
	}
	return suggestions
}

func extractRecipeFields(response *fetchers.Response, recipe ExtractionRecipe) map[string]ExtractedField {
	if len(recipe.Fields) == 0 {
		return nil
	}
	fields := make(map[string]ExtractedField, len(recipe.Fields))
	for _, field := range recipe.Fields {
		name := strings.TrimSpace(field.Name)
		if name == "" {
			name = strings.TrimSpace(field.Selector)
		}
		if name == "" {
			continue
		}
		values, err := extractRecipeField(response, field)
		fields[name] = extractedFieldResult(field, values, err)
	}
	return fields
}

func extractRecipeField(response *fetchers.Response, field ExtractionField) ([]string, error) {
	selector := strings.TrimSpace(field.Selector)
	if selector == "" {
		return nil, fmt.Errorf("selector is required")
	}
	var values []string
	var err error
	switch field.Type {
	case "", SelectorCSS:
		selection := response.CSS(selector)
		values = selection.GetAll().Strings()
	case SelectorXPath:
		values, err = extractXPathValues(response.Text(), selector)
	case SelectorJSON:
		values, err = extractJSONValues(response, selector, field.Multiple)
	default:
		return nil, fmt.Errorf("unsupported selector type %q", field.Type)
	}
	if err != nil {
		return nil, err
	}
	return transformRecipeValues(values, field), nil
}

func extractedFieldResult(field ExtractionField, values []string, err error) ExtractedField {
	result := ExtractedField{Type: field.Type, Selector: field.Selector}
	if result.Type == "" {
		result.Type = SelectorCSS
	}
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(values) == 0 {
		result.Error = "selector matched no values"
		return result
	}
	if field.Multiple {
		result.Values = append([]string(nil), values...)
		result.Value = values[0]
		return result
	}
	result.Value = values[0]
	if len(values) > 1 {
		result.Values = append([]string(nil), values...)
	}
	return result
}

func extractXPathValues(body string, selector string) ([]string, error) {
	document, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	nodes, err := htmlquery.QueryAll(document, selector)
	if err != nil {
		return nil, fmt.Errorf("invalid xpath %q: %w", selector, err)
	}
	values := make([]string, 0, len(nodes))
	for _, node := range nodes {
		value := normalizeRecipeValue(xpathNodeText(node))
		if value != "" {
			values = append(values, value)
		}
	}
	return values, nil
}

func xpathNodeText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}
	return htmlquery.InnerText(node)
}

func extractJSONValues(response *fetchers.Response, selector string, multiple bool) ([]string, error) {
	var body any
	if err := response.DecodeJSON(&body); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}
	matches := queryJSONPath(body, selector)
	values := make([]string, 0, len(matches))
	for _, match := range matches {
		values = append(values, jsonValueStrings(match, multiple)...)
	}
	return values, nil
}

func queryJSONPath(data any, selector string) []any {
	parts := splitJSONPath(selector)
	if len(parts) == 0 {
		return nil
	}
	return queryJSONParts(data, parts)
}

func splitJSONPath(selector string) []string {
	rawParts := strings.Split(selector, "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func queryJSONParts(data any, parts []string) []any {
	if len(parts) == 0 {
		return []any{data}
	}
	part := parts[0]
	rest := parts[1:]
	switch value := data.(type) {
	case map[string]any:
		child, ok := value[part]
		if !ok {
			return nil
		}
		return queryJSONParts(child, rest)
	case []any:
		if index, err := strconv.Atoi(part); err == nil {
			if index < 0 || index >= len(value) {
				return nil
			}
			return queryJSONParts(value[index], rest)
		}
		var matches []any
		for _, child := range value {
			matches = append(matches, queryJSONParts(child, parts)...)
		}
		return matches
	default:
		return nil
	}
}

func jsonValueStrings(value any, multiple bool) []string {
	if multiple {
		if values, ok := value.([]any); ok {
			output := make([]string, 0, len(values))
			for _, item := range values {
				if text := jsonValueString(item); text != "" {
					output = append(output, text)
				}
			}
			return output
		}
	}
	if text := jsonValueString(value); text != "" {
		return []string{text}
	}
	return nil
}

func jsonValueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return normalizeRecipeValue(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(typed)
	default:
		body, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(body)
	}
}

func transformRecipeValues(values []string, field ExtractionField) []string {
	if len(values) == 0 {
		return nil
	}
	transformed := make([]string, 0, len(values))
	for _, value := range values {
		if field.HTMLToText {
			value = recipeHTMLToText(value)
		}
		value = normalizeRecipeValue(value)
		if value == "" {
			continue
		}
		if field.Prefix != "" {
			value = field.Prefix + value
		}
		transformed = append(transformed, value)
	}
	return transformed
}

func recipeHTMLToText(value string) string {
	document, err := html.Parse(strings.NewReader(value))
	if err != nil {
		return normalizeRecipeValue(value)
	}
	var builder strings.Builder
	collectRecipeHTMLText(&builder, document)
	return normalizeRecipeValue(builder.String())
}

func collectRecipeHTMLText(builder *strings.Builder, node *html.Node) {
	if node == nil {
		return
	}
	if node.Type == html.ElementNode && (strings.EqualFold(node.Data, "script") || strings.EqualFold(node.Data, "style")) {
		return
	}
	if node.Type == html.TextNode {
		builder.WriteByte(' ')
		builder.WriteString(node.Data)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		collectRecipeHTMLText(builder, child)
	}
}

func normalizeRecipeValue(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
