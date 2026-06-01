package spiders

import (
	"net/url"
	"path"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

var DefaultIgnoredExtensions = map[string]struct{}{
	"7z": {}, "7zip": {}, "bz2": {}, "rar": {}, "tar": {}, "tar.gz": {}, "xz": {}, "zip": {},
	"mng": {}, "pct": {}, "bmp": {}, "gif": {}, "jpg": {}, "jpeg": {}, "png": {}, "pst": {}, "psp": {}, "tif": {}, "tiff": {}, "ai": {}, "drw": {}, "dxf": {}, "eps": {}, "ps": {}, "svg": {}, "cdr": {}, "ico": {}, "webp": {},
	"mp3": {}, "wma": {}, "ogg": {}, "wav": {}, "ra": {}, "aac": {}, "mid": {}, "au": {}, "aiff": {},
	"3gp": {}, "asf": {}, "asx": {}, "avi": {}, "mov": {}, "mp4": {}, "mpg": {}, "qt": {}, "rm": {}, "swf": {}, "wmv": {}, "m4a": {}, "m4v": {}, "flv": {}, "webm": {},
	"xls": {}, "xlsm": {}, "xlsx": {}, "xltm": {}, "xltx": {}, "potm": {}, "potx": {}, "ppt": {}, "pptm": {}, "pptx": {}, "pps": {}, "doc": {}, "docb": {}, "docm": {}, "docx": {}, "dotm": {}, "dotx": {}, "odt": {}, "ods": {}, "odg": {}, "odp": {},
	"css": {}, "pdf": {}, "exe": {}, "bin": {}, "rss": {}, "dmg": {}, "iso": {}, "apk": {}, "jar": {}, "sh": {}, "rb": {}, "js": {}, "hta": {}, "bat": {}, "cpl": {}, "msi": {}, "msp": {}, "py": {},
}

type LinkProcessFunc func(string) (string, bool)

type LinkExtractorOptions struct {
	Allow               []string
	Deny                []string
	AllowDomains        []string
	DenyDomains         []string
	RestrictCSS         []string
	RestrictXPath       []string
	Tags                []string
	Attrs               []string
	DisableCanonicalize bool
	DisableStrip        bool
	KeepFragment        bool
	DenyExtensions      []string
	Process             LinkProcessFunc
}

type LinkExtractor struct {
	allow          []*regexp.Regexp
	deny           []*regexp.Regexp
	allowDomains   []string
	denyDomains    []string
	restrictCSS    []string
	restrictXPath  []string
	tags           map[string]struct{}
	attrs          map[string]struct{}
	canonicalize   bool
	strip          bool
	keepFragment   bool
	denyExtensions map[string]struct{}
	process        LinkProcessFunc
}

func NewLinkExtractor(opts LinkExtractorOptions) (*LinkExtractor, error) {
	allow, err := compileLinkPatterns(opts.Allow)
	if err != nil {
		return nil, err
	}
	deny, err := compileLinkPatterns(opts.Deny)
	if err != nil {
		return nil, err
	}
	tags := opts.Tags
	if len(tags) == 0 {
		tags = []string{"a", "area"}
	}
	attrs := opts.Attrs
	if len(attrs) == 0 {
		attrs = []string{"href"}
	}
	denyExtensions := make(map[string]struct{})
	if opts.DenyExtensions == nil {
		for ext := range DefaultIgnoredExtensions {
			denyExtensions[ext] = struct{}{}
		}
	} else {
		for _, ext := range opts.DenyExtensions {
			ext = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), "."))
			if ext != "" {
				denyExtensions[ext] = struct{}{}
			}
		}
	}
	process := opts.Process
	if process == nil {
		process = defaultLinkProcess
	}
	return &LinkExtractor{
		allow:          allow,
		deny:           deny,
		allowDomains:   normalizeLinkDomains(opts.AllowDomains),
		denyDomains:    normalizeLinkDomains(opts.DenyDomains),
		restrictCSS:    append([]string(nil), opts.RestrictCSS...),
		restrictXPath:  append([]string(nil), opts.RestrictXPath...),
		tags:           stringSet(tags),
		attrs:          stringSet(attrs),
		canonicalize:   !opts.DisableCanonicalize,
		strip:          !opts.DisableStrip,
		keepFragment:   opts.KeepFragment,
		denyExtensions: denyExtensions,
		process:        process,
	}, nil
}

func (e *LinkExtractor) Extract(response Response) ([]string, error) {
	if e == nil {
		return nil, nil
	}
	scopes, err := e.scopes(response)
	if err != nil {
		return nil, err
	}
	baseURL := linkDocumentBaseURL(response.URL(), string(response.Body()))
	seen := map[string]struct{}{}
	links := make([]string, 0)
	for _, scope := range scopes {
		extracted, err := e.extractScope(baseURL, scope)
		if err != nil {
			return nil, err
		}
		for _, link := range extracted {
			if _, ok := seen[link]; ok {
				continue
			}
			seen[link] = struct{}{}
			links = append(links, link)
		}
	}
	return links, nil
}

func (e *LinkExtractor) Matches(rawURL string) bool {
	_, ok := e.MatchURL(rawURL)
	return ok
}

// MatchURL returns the processed, canonical URL that made rawURL match.
func (e *LinkExtractor) MatchURL(rawURL string) (string, bool) {
	if e == nil {
		return "", false
	}
	return e.MatchURLFrom(matchCandidateBaseURL(rawURL, e.strip), rawURL)
}

// MatchURLFrom returns the processed, canonical URL that made rawURL match after
// resolving relative candidates against baseURL.
func (e *LinkExtractor) MatchURLFrom(baseURL, rawURL string) (string, bool) {
	if e == nil {
		return "", false
	}
	if baseURL == "" {
		baseURL = matchCandidateBaseURL(rawURL, e.strip)
	}
	result := e.prepareCandidateDiagnostic(baseURL, rawURL)
	if !result.ok {
		return "", false
	}
	return result.candidate.url, true
}

func matchCandidateBaseURL(rawURL string, strip bool) string {
	if strip {
		return strings.TrimSpace(rawURL)
	}
	return rawURL
}

func (e *LinkExtractor) scopes(response Response) ([]string, error) {
	wholeResponse := string(response.Body())
	if len(e.restrictCSS) == 0 && len(e.restrictXPath) == 0 {
		return []string{wholeResponse}, nil
	}

	scopes := make([]string, 0, len(e.restrictXPath)+len(e.restrictCSS))
	for _, expr := range e.restrictXPath {
		htmlText, err := response.XPath(expr).HTML()
		if err != nil {
			return nil, err
		}
		scopes = appendNonEmptyScope(scopes, htmlText)
	}
	for _, selector := range e.restrictCSS {
		htmlText, err := response.CSS(selector).HTML()
		if err != nil {
			return nil, err
		}
		scopes = appendNonEmptyScope(scopes, htmlText)
	}
	if len(scopes) == 0 {
		return []string{wholeResponse}, nil
	}
	return scopes, nil
}

func appendNonEmptyScope(scopes []string, htmlText string) []string {
	if htmlText == "" {
		return scopes
	}
	return append(scopes, htmlText)
}

func linkDocumentBaseURL(responseURL, body string) string {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return responseURL
	}
	rawBase := firstBaseHref(root)
	if rawBase == "" {
		return responseURL
	}
	resolved, err := resolveURL(responseURL, strings.TrimSpace(rawBase))
	if err != nil {
		return responseURL
	}
	return resolved
}

func firstBaseHref(root *html.Node) string {
	var walk func(*html.Node) string
	walk = func(node *html.Node) string {
		if node == nil {
			return ""
		}
		if node.Type == html.ElementNode && strings.EqualFold(node.Data, "base") {
			for _, attr := range node.Attr {
				if strings.EqualFold(attr.Key, "href") {
					return attr.Val
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if value := walk(child); value != "" {
				return value
			}
		}
		return ""
	}
	return walk(root)
}

func (e *LinkExtractor) extractScope(baseURL, body string) ([]string, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	candidates := collectLinkAttributeCandidates(root, e.tags, e.attrs)
	links := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		link, ok, err := e.prepareURL(baseURL, candidate.raw)
		if err != nil {
			return nil, err
		}
		if ok {
			links = append(links, link)
		}
	}
	return links, nil
}

type linkAttributeCandidate struct {
	tag  string
	attr string
	raw  string
}

func collectLinkAttributeCandidates(root *html.Node, tags, attrs map[string]struct{}) []linkAttributeCandidate {
	candidates := make([]linkAttributeCandidate, 0)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if node.Type == html.ElementNode {
			tag := strings.ToLower(node.Data)
			if _, ok := tags[tag]; ok {
				for _, rawAttr := range node.Attr {
					attr := strings.ToLower(rawAttr.Key)
					if _, ok := attrs[attr]; !ok {
						continue
					}
					candidates = append(candidates, linkAttributeCandidate{tag: tag, attr: attr, raw: rawAttr.Val})
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return candidates
}

type linkURLCandidate struct {
	baseURL  string
	raw      string
	resolved string
}

type preparedLinkCandidate struct {
	candidate linkURLCandidate
	url       string
}

type linkDropReason string

const (
	linkDropNone             linkDropReason = ""
	linkDropEmptyRaw         linkDropReason = "empty_raw"
	linkDropInvalidRaw       linkDropReason = "invalid_raw"
	linkDropProcessRejected  linkDropReason = "process_rejected"
	linkDropInvalidProcessed linkDropReason = "invalid_processed"
	linkDropInvalidCanonical linkDropReason = "invalid_canonical"
	linkDropFiltered         linkDropReason = "filtered"
)

type preparedLinkCandidateResult struct {
	candidate preparedLinkCandidate
	ok        bool
	err       error
	reason    linkDropReason
}

func droppedLinkCandidate(reason linkDropReason) preparedLinkCandidateResult {
	return preparedLinkCandidateResult{reason: reason}
}

func newLinkURLCandidate(baseURL, raw string, strip bool) (linkURLCandidate, bool, error) {
	if strip {
		raw = strings.TrimSpace(raw)
	}
	if raw == "" {
		return linkURLCandidate{}, false, nil
	}
	resolved, err := resolveURL(baseURL, raw)
	if err != nil {
		return linkURLCandidate{}, false, err
	}
	return linkURLCandidate{baseURL: baseURL, raw: raw, resolved: resolved}, true, nil
}

func (candidate linkURLCandidate) applyProcess(process LinkProcessFunc) (preparedLinkCandidate, bool, error) {
	processed, ok := process(candidate.resolved)
	if !ok || processed == "" {
		return preparedLinkCandidate{}, false, nil
	}
	resolved, err := resolveURL(candidate.baseURL, processed)
	if err != nil {
		return preparedLinkCandidate{}, false, err
	}
	return preparedLinkCandidate{candidate: candidate, url: resolved}, true, nil
}

func (candidate preparedLinkCandidate) canonicalURL(keepFragment bool) (preparedLinkCandidate, error) {
	canonical, err := canonicalizeLinkURL(candidate.url, keepFragment)
	if err != nil {
		return preparedLinkCandidate{}, err
	}
	candidate.url = canonical
	return candidate, nil
}

func (e *LinkExtractor) prepareURL(baseURL, raw string) (string, bool, error) {
	candidate, ok, err := e.prepareCandidate(baseURL, raw)
	if !ok || err != nil {
		return "", ok, err
	}
	return candidate.url, true, nil
}

func (e *LinkExtractor) prepareCandidate(baseURL, raw string) (preparedLinkCandidate, bool, error) {
	result := e.prepareCandidateDiagnostic(baseURL, raw)
	return result.candidate, result.ok, result.err
}

func (e *LinkExtractor) prepareCandidateDiagnostic(baseURL, raw string) preparedLinkCandidateResult {
	return e.candidateConfig().prepare(baseURL, raw)
}

func defaultLinkProcess(value string) (string, bool) { return value, true }

type linkCandidateConfig struct {
	strip        bool
	process      LinkProcessFunc
	canonicalize bool
	keepFragment bool
	passes       func(string) bool
}

func (e *LinkExtractor) candidateConfig() linkCandidateConfig {
	return linkCandidateConfig{
		strip:        e.strip,
		process:      e.process,
		canonicalize: e.canonicalize,
		keepFragment: e.keepFragment,
		passes:       e.urlPasses,
	}
}

func (config linkCandidateConfig) prepare(baseURL, raw string) preparedLinkCandidateResult {
	candidate, ok, err := newLinkURLCandidate(baseURL, raw, config.strip)
	if err != nil {
		return droppedLinkCandidate(linkDropInvalidRaw)
	}
	if !ok {
		return droppedLinkCandidate(linkDropEmptyRaw)
	}
	prepared, ok, err := candidate.applyProcess(config.process)
	if err != nil {
		return droppedLinkCandidate(linkDropInvalidProcessed)
	}
	if !ok {
		return droppedLinkCandidate(linkDropProcessRejected)
	}
	if config.canonicalize {
		prepared, err = prepared.canonicalURL(config.keepFragment)
		if err != nil {
			return droppedLinkCandidate(linkDropInvalidCanonical)
		}
	}
	if !config.passes(prepared.url) {
		return droppedLinkCandidate(linkDropFiltered)
	}
	return preparedLinkCandidateResult{candidate: prepared, ok: true, reason: linkDropNone}
}

func (e *LinkExtractor) urlPasses(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if !linkHasSupportedLocator(parsed) {
		return false
	}
	if ext := linkExtension(parsed.Path); ext != "" {
		if _, denied := e.denyExtensions[ext]; denied {
			return false
		}
	}
	if len(e.allow) > 0 && !anyPatternMatches(e.allow, rawURL) {
		return false
	}
	if len(e.deny) > 0 && anyPatternMatches(e.deny, rawURL) {
		return false
	}
	if len(e.allowDomains) > 0 || len(e.denyDomains) > 0 {
		host := strings.ToLower(strings.TrimSuffix(parsed.Hostname(), "."))
		if len(e.allowDomains) > 0 && !domainInList(host, e.allowDomains) {
			return false
		}
		if len(e.denyDomains) > 0 && domainInList(host, e.denyDomains) {
			return false
		}
	}
	return true
}

func linkHasSupportedLocator(parsed *url.URL) bool {
	scheme := strings.ToLower(parsed.Scheme)
	switch scheme {
	case "http", "https":
		return parsed.Host != ""
	case "file":
		return true
	default:
		return false
	}
}

func compileLinkPatterns(patterns []string) ([]*regexp.Regexp, error) {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, re)
	}
	return compiled, nil
}

func normalizeLinkDomains(domains []string) []string {
	out := make([]string, 0, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
		if domain != "" {
			out = append(out, domain)
		}
	}
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func canonicalizeLinkURL(rawURL string, keepFragment bool) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	normalizeParsedURLHost(parsed)
	if parsed.Path != "" {
		cleaned := path.Clean(parsed.Path)
		if strings.HasSuffix(parsed.Path, "/") && !strings.HasSuffix(cleaned, "/") {
			cleaned += "/"
		}
		parsed.Path = cleaned
	}
	parsed.RawQuery = parsed.Query().Encode()
	if !keepFragment {
		parsed.Fragment = ""
	}
	return parsed.String(), nil
}

func anyPatternMatches(patterns []*regexp.Regexp, value string) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

func domainInList(host string, domains []string) bool {
	for _, domain := range domains {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return true
		}
	}
	return false
}

func linkExtension(pathValue string) string {
	last := linkLastPathSegment(pathValue)
	last = strings.ToLower(stripLinkPathParameters(last))
	if strings.HasSuffix(last, ".tar.gz") {
		return "tar.gz"
	}
	idx := strings.LastIndex(last, ".")
	if idx == -1 || idx == len(last)-1 {
		return ""
	}
	return last[idx+1:]
}

func linkLastPathSegment(pathValue string) string {
	trimmed := strings.Trim(pathValue, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	return parts[len(parts)-1]
}

func stripLinkPathParameters(segment string) string {
	beforeParams, _, _ := strings.Cut(segment, ";")
	return beforeParams
}

func sortedLinkMapKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
