package parser

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/TrebuchetDynamics/goscrapling/core/storage"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type ParseOptions struct {
	URL   string
	Store Store
}

type SelectorOptions struct {
	Identifier string
	Adaptive   bool
	AutoSave   bool
	Percentage float64
	Domain     string
}

type Document struct {
	root   *html.Node
	query  *goquery.Document
	url    string
	domain string
	store  Store
}

func Parse(r io.Reader, opts ParseOptions) (*Document, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return &Document{
		root:   root,
		query:  goquery.NewDocumentFromNode(root),
		url:    opts.URL,
		domain: storage.AdaptiveDomain(opts.URL),
		store:  opts.Store,
	}, nil
}

func (d *Document) CSS(selector string) Selection {
	if d == nil || d.query == nil {
		return Selection{}
	}

	selection, err := d.css(selector)
	if err != nil {
		return Selection{}
	}
	return selection
}

func (d *Document) SelectCSS(ctx context.Context, selector string, opts SelectorOptions) (Selection, error) {
	if d == nil || d.query == nil {
		return Selection{}, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	selector = strings.TrimSpace(selector)
	minScore, err := minScoreFromPercentage(opts.Percentage)
	if err != nil {
		return Selection{}, err
	}

	branches := selectorBranches(selector, opts.Identifier)
	elements := make([]*Element, 0)
	for _, branch := range branches {
		selection, err := d.selectCSSBranch(ctx, branch, opts, minScore)
		if err != nil {
			return Selection{}, err
		}
		elements = append(elements, selection.elements...)
	}

	return Selection{elements: elements}, nil
}

func (d *Document) selectCSSBranch(ctx context.Context, selector string, opts SelectorOptions, minScore float64) (Selection, error) {
	baseSelector, extraction, err := parseCSSExtraction(selector)
	if err != nil {
		return Selection{}, err
	}

	matcher, err := cascadia.Compile(baseSelector)
	if err != nil {
		return Selection{}, fmt.Errorf("%w %q: %v", ErrInvalidSelector, selector, err)
	}

	key := d.selectorKey(selector, opts)
	selection := d.cssWithMatcher(matcher).applyExtraction(extraction)
	if selection.Len() > 0 {
		if opts.AutoSave {
			if d.store == nil {
				return Selection{}, ErrMissingStore
			}
			first, _ := selection.First()
			if err := d.store.Save(ctx, key, storage.FingerprintNode(first.node)); err != nil {
				return Selection{}, err
			}
		}
		return selection, nil
	}

	if !opts.Adaptive {
		if opts.AutoSave && d.store == nil {
			return Selection{}, ErrMissingStore
		}
		return Selection{}, nil
	}
	if d.store == nil {
		return Selection{}, ErrMissingStore
	}

	match, ok, err := d.relocateWithKey(ctx, key, minScore)
	if err != nil || !ok {
		return Selection{}, err
	}

	if opts.AutoSave {
		if err := d.store.Save(ctx, key, storage.FingerprintNode(match.Element.node)); err != nil {
			return Selection{}, err
		}
	}
	return Selection{elements: []*Element{match.Element}}.applyExtraction(extraction), nil
}

func (d *Document) css(selector string) (Selection, error) {
	branches := selectorBranches(selector, "")
	var combined Selection

	for _, branch := range branches {
		selection, err := d.cssBranch(branch)
		if err != nil {
			return Selection{}, err
		}
		combined = combineSelections(combined, selection)
	}

	return combined, nil
}

func (d *Document) cssBranch(selector string) (Selection, error) {
	baseSelector, extraction, err := parseCSSExtraction(selector)
	if err != nil {
		return Selection{}, err
	}

	matcher, err := cascadia.Compile(baseSelector)
	if err != nil {
		return Selection{}, fmt.Errorf("%w %q: %v", ErrInvalidSelector, selector, err)
	}

	return d.cssWithMatcher(matcher).applyExtraction(extraction), nil
}

func (d *Document) cssWithMatcher(matcher cascadia.Selector) Selection {
	var elements []*Element
	d.query.FindMatcher(matcher).Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			elements = append(elements, &Element{doc: d, node: node})
		}
	})
	return Selection{elements: elements}
}

type cssExtractionKind int

const (
	cssExtractionElement cssExtractionKind = iota
	cssExtractionText
	cssExtractionAttr
)

type cssExtraction struct {
	kind cssExtractionKind
	attr string
}

func (s Selection) applyExtraction(extraction cssExtraction) Selection {
	switch extraction.kind {
	case cssExtractionText:
		return s.withTextValues()
	case cssExtractionAttr:
		return s.withAttrValues(extraction.attr)
	default:
		return s
	}
}

func parseCSSExtraction(selector string) (string, cssExtraction, error) {
	selector = strings.TrimSpace(selector)
	pseudoIndex := topLevelPseudoElementIndex(selector)
	if pseudoIndex == -1 {
		return selector, cssExtraction{kind: cssExtractionElement}, nil
	}

	baseSelector := strings.TrimSpace(selector[:pseudoIndex])
	if baseSelector == "" {
		baseSelector = "*"
	}
	pseudo := strings.TrimSpace(selector[pseudoIndex:])
	if pseudo == "::text" {
		return baseSelector, cssExtraction{kind: cssExtractionText}, nil
	}

	if strings.HasPrefix(pseudo, "::attr(") && strings.HasSuffix(pseudo, ")") {
		name := strings.TrimSpace(pseudo[len("::attr(") : len(pseudo)-1])
		name = strings.Trim(name, `"'`)
		name = strings.TrimSpace(name)
		if name == "" || strings.ContainsAny(name, " \t\r\n") {
			return "", cssExtraction{}, fmt.Errorf("%w %q: invalid ::attr() argument", ErrInvalidSelector, selector)
		}
		return baseSelector, cssExtraction{kind: cssExtractionAttr, attr: name}, nil
	}

	return "", cssExtraction{}, fmt.Errorf("%w %q: unknown pseudo-element", ErrInvalidSelector, selector)
}

func topLevelPseudoElementIndex(selector string) int {
	quote := byte(0)
	parenDepth := 0
	bracketDepth := 0
	escaped := false
	pseudoIndex := -1

	for i := 0; i < len(selector)-1; i++ {
		ch := selector[i]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			continue
		}

		switch ch {
		case '\'', '"':
			quote = ch
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case ':':
			if selector[i+1] == ':' && parenDepth == 0 && bracketDepth == 0 {
				pseudoIndex = i
			}
		}
	}

	return pseudoIndex
}

func selectorBranches(selector string, identifier string) []string {
	identifier = strings.TrimSpace(identifier)
	if identifier != "" {
		return []string{selector}
	}

	parts := splitTopLevelCommas(selector)
	branches := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			branches = append(branches, part)
		}
	}
	if len(branches) == 0 {
		return []string{selector}
	}
	return branches
}

func splitTopLevelCommas(selector string) []string {
	var parts []string
	var builder strings.Builder
	var quote rune
	parenDepth := 0
	bracketDepth := 0
	escaped := false

	for _, r := range selector {
		if escaped {
			builder.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			builder.WriteRune(r)
			escaped = true
			continue
		}
		if quote != 0 {
			builder.WriteRune(r)
			if r == quote {
				quote = 0
			}
			continue
		}

		switch r {
		case '\'', '"':
			quote = r
			builder.WriteRune(r)
		case '(':
			parenDepth++
			builder.WriteRune(r)
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			builder.WriteRune(r)
		case '[':
			bracketDepth++
			builder.WriteRune(r)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			builder.WriteRune(r)
		case ',':
			if parenDepth == 0 && bracketDepth == 0 {
				parts = append(parts, builder.String())
				builder.Reset()
				continue
			}
			builder.WriteRune(r)
		default:
			builder.WriteRune(r)
		}
	}

	parts = append(parts, builder.String())
	return parts
}

func minScoreFromPercentage(percentage float64) (float64, error) {
	if percentage == 0 {
		return storage.DefaultMinScore, nil
	}
	if percentage < 0 || percentage > 100 {
		return 0, fmt.Errorf("%w: %v", ErrInvalidPercentage, percentage)
	}
	return percentage / 100, nil
}

func (d *Document) selectorKey(selector string, opts SelectorOptions) Key {
	identifier := strings.TrimSpace(opts.Identifier)
	if identifier == "" {
		identifier = selector
	}
	return Key{
		Domain:     selectorDomain(d.domain, opts.Domain),
		Identifier: identifier,
	}
}

func selectorDomain(defaultValue string, override string) string {
	override = strings.TrimSpace(override)
	if override == "" {
		return defaultValue
	}

	if parsed, err := url.Parse(override); err == nil && parsed.Host != "" {
		host := strings.TrimSpace(strings.ToLower(parsed.Hostname()))
		if host != "" {
			return host
		}
	}

	if host, _, err := net.SplitHostPort(override); err == nil {
		override = host
	}
	override = strings.TrimSpace(strings.ToLower(override))
	if override == "" {
		return storage.DefaultDomain
	}
	return override
}

func (d *Document) Save(ctx context.Context, element *Element, identifier string) error {
	if d == nil || d.store == nil {
		return ErrMissingStore
	}
	if element == nil || element.node == nil {
		return ErrNilElement
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return ErrEmptyIdentifier
	}

	return d.store.Save(ctx, Key{Domain: d.domain, Identifier: identifier}, storage.FingerprintNode(element.node))
}

func (d *Document) Retrieve(ctx context.Context, identifier string) (*Element, bool, error) {
	match, ok, err := d.Relocate(ctx, identifier)
	if err != nil || !ok {
		return nil, ok, err
	}
	return match.Element, true, nil
}

func (d *Document) Relocate(ctx context.Context, identifier string) (Match, bool, error) {
	if d == nil || d.store == nil {
		return Match{}, false, ErrMissingStore
	}

	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return Match{}, false, ErrEmptyIdentifier
	}

	return d.relocateWithKey(ctx, Key{Domain: d.domain, Identifier: identifier}, storage.DefaultMinScore)
}

func (d *Document) relocateWithKey(ctx context.Context, key Key, minScore float64) (Match, bool, error) {
	if d == nil || d.store == nil {
		return Match{}, false, ErrMissingStore
	}
	if ctx == nil {
		ctx = context.Background()
	}

	key.Identifier = strings.TrimSpace(key.Identifier)
	if key.Identifier == "" {
		return Match{}, false, ErrEmptyIdentifier
	}
	key.Domain = strings.TrimSpace(key.Domain)
	if key.Domain == "" {
		key.Domain = storage.DefaultDomain
	}

	target, ok, err := d.store.Load(ctx, key)
	if err != nil || !ok {
		return Match{}, false, err
	}

	var best Match
	found := false
	if d.query == nil {
		return Match{}, false, nil
	}
	d.query.Find("*").Each(func(_ int, selection *goquery.Selection) {
		for _, node := range selection.Nodes {
			candidate := &Element{doc: d, node: node}
			score := storage.ScoreFingerprint(storage.FingerprintNode(node), target)
			if score < minScore {
				continue
			}
			if !found || score > best.Score {
				best = Match{Element: candidate, Score: score}
				found = true
			}
		}
	})

	return best, found, nil
}
