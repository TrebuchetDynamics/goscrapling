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
		process = func(value string) (string, bool) { return value, true }
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
	seen := map[string]struct{}{}
	links := make([]string, 0)
	for _, scope := range scopes {
		extracted, err := e.extractScope(response.URL(), scope)
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
	if e == nil {
		return false
	}
	if e.canonicalize {
		canonical, err := canonicalizeLinkURL(rawURL, e.keepFragment)
		if err != nil {
			return false
		}
		rawURL = canonical
	}
	return e.urlPasses(rawURL)
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

func (e *LinkExtractor) extractScope(baseURL, body string) ([]string, error) {
	root, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	links := make([]string, 0)
	var walk func(*html.Node) error
	walk = func(node *html.Node) error {
		if node == nil {
			return nil
		}
		if node.Type == html.ElementNode {
			if _, ok := e.tags[strings.ToLower(node.Data)]; ok {
				for _, attr := range node.Attr {
					if _, ok := e.attrs[strings.ToLower(attr.Key)]; !ok {
						continue
					}
					link, ok, err := e.prepareURL(baseURL, attr.Val)
					if err != nil {
						return err
					}
					if ok {
						links = append(links, link)
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root); err != nil {
		return nil, err
	}
	return links, nil
}

func (e *LinkExtractor) prepareURL(baseURL, raw string) (string, bool, error) {
	if e.strip {
		raw = strings.TrimSpace(raw)
	}
	if raw == "" {
		return "", false, nil
	}
	resolved, err := resolveURL(baseURL, raw)
	if err != nil {
		return "", false, err
	}
	processed, ok := e.process(resolved)
	if !ok || processed == "" {
		return "", false, nil
	}
	if e.canonicalize {
		processed, err = canonicalizeLinkURL(processed, e.keepFragment)
		if err != nil {
			return "", false, nil
		}
	}
	if !e.urlPasses(processed) {
		return "", false, nil
	}
	return processed, true, nil
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
	parsed.Host = strings.ToLower(parsed.Host)
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
	_, last, ok := strings.Cut(strings.TrimLeft(pathValue, "/"), "/")
	if ok {
		parts := strings.Split(strings.Trim(pathValue, "/"), "/")
		last = parts[len(parts)-1]
	} else {
		last = strings.Trim(pathValue, "/")
	}
	last = strings.ToLower(last)
	if strings.HasSuffix(last, ".tar.gz") {
		return "tar.gz"
	}
	idx := strings.LastIndex(last, ".")
	if idx == -1 || idx == len(last)-1 {
		return ""
	}
	return last[idx+1:]
}

func sortedLinkMapKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
