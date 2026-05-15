package translator

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrInvalidSelector = errors.New("goscrapling: invalid selector")

func CSSToXPath(selector string) (string, error) {
	branches := selectorBranches(selector, "")
	xpaths := make([]string, 0, len(branches))
	for _, branch := range branches {
		xpath, err := cssBranchToXPath(branch)
		if err != nil {
			return "", err
		}
		xpaths = append(xpaths, xpath)
	}
	return strings.Join(xpaths, " | "), nil
}

func cssBranchToXPath(selector string) (string, error) {
	baseSelector, extraction, err := parseCSSExtraction(selector)
	if err != nil {
		return "", err
	}
	steps, err := parseCSSSteps(baseSelector)
	if err != nil {
		return "", err
	}
	if len(steps) == 0 {
		return "", fmt.Errorf("%w %q: empty selector", ErrInvalidSelector, selector)
	}

	var builder strings.Builder
	for index, step := range steps {
		xpathStep, err := cssSimpleSelectorToXPath(step.selector)
		if err != nil {
			return "", err
		}
		switch {
		case index == 0:
			builder.WriteString("//")
		case step.combinator == cssCombinatorChild:
			builder.WriteString("/")
		default:
			builder.WriteString("//")
		}
		builder.WriteString(xpathStep)
	}

	switch extraction.kind {
	case cssExtractionText:
		builder.WriteString("/text()")
	case cssExtractionAttr:
		builder.WriteString("/@")
		builder.WriteString(extraction.attr)
	}
	return builder.String(), nil
}

type cssCombinator int

const (
	cssCombinatorDescendant cssCombinator = iota
	cssCombinatorChild
)

type cssStep struct {
	combinator cssCombinator
	selector   string
}

func parseCSSSteps(selector string) ([]cssStep, error) {
	selector = strings.TrimSpace(selector)
	if selector == "" {
		return nil, fmt.Errorf("%w: empty selector", ErrInvalidSelector)
	}

	var steps []cssStep
	var builder strings.Builder
	combinator := cssCombinatorDescendant
	quote := rune(0)
	bracketDepth := 0
	escaped := false

	flush := func() {
		part := strings.TrimSpace(builder.String())
		if part != "" {
			steps = append(steps, cssStep{combinator: combinator, selector: part})
			builder.Reset()
			combinator = cssCombinatorDescendant
		}
	}

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
		case '[':
			bracketDepth++
			builder.WriteRune(r)
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			builder.WriteRune(r)
		case '>':
			if bracketDepth == 0 {
				flush()
				combinator = cssCombinatorChild
				continue
			}
			builder.WriteRune(r)
		default:
			if unicode.IsSpace(r) && bracketDepth == 0 {
				flush()
				continue
			}
			builder.WriteRune(r)
		}
	}
	flush()

	return steps, nil
}

func cssSimpleSelectorToXPath(selector string) (string, error) {
	parser := cssSimpleParser{input: selector}
	tag := "*"
	if parser.hasIdentStart() || parser.peek() == '*' {
		value := parser.readIdent()
		if value == "" {
			value = parser.readRune()
		}
		if value != "*" {
			tag = strings.ToLower(value)
		}
	}

	var predicates []string
	for !parser.done() {
		switch parser.peek() {
		case '.':
			parser.advance()
			className := parser.readIdent()
			if className == "" {
				return "", fmt.Errorf("%w %q: empty class selector", ErrInvalidSelector, selector)
			}
			predicates = append(predicates, "contains(concat(' ', normalize-space(@class), ' '), "+xpathLiteral(" "+className+" ")+")")
		case '#':
			parser.advance()
			id := parser.readIdent()
			if id == "" {
				return "", fmt.Errorf("%w %q: empty id selector", ErrInvalidSelector, selector)
			}
			predicates = append(predicates, "@id="+xpathLiteral(id))
		case '[':
			predicate, err := parser.readAttributePredicate()
			if err != nil {
				return "", fmt.Errorf("%w %q: %v", ErrInvalidSelector, selector, err)
			}
			predicates = append(predicates, predicate)
		default:
			return "", fmt.Errorf("%w %q: unsupported selector syntax", ErrInvalidSelector, selector)
		}
	}

	if len(predicates) == 0 {
		return tag, nil
	}
	return tag + "[" + strings.Join(predicates, " and ") + "]", nil
}

type cssSimpleParser struct {
	input string
	pos   int
}

func (p *cssSimpleParser) done() bool {
	return p.pos >= len(p.input)
}

func (p *cssSimpleParser) peek() byte {
	if p.done() {
		return 0
	}
	return p.input[p.pos]
}

func (p *cssSimpleParser) advance() {
	if !p.done() {
		p.pos++
	}
}

func (p *cssSimpleParser) readRune() string {
	if p.done() {
		return ""
	}
	value := p.input[p.pos : p.pos+1]
	p.pos++
	return value
}

func (p *cssSimpleParser) hasIdentStart() bool {
	if p.done() {
		return false
	}
	ch := p.peek()
	return isCSSIdentChar(ch)
}

func (p *cssSimpleParser) readIdent() string {
	start := p.pos
	for !p.done() && isCSSIdentChar(p.peek()) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *cssSimpleParser) readAttributePredicate() (string, error) {
	p.advance()
	start := p.pos
	quote := byte(0)
	for !p.done() {
		ch := p.peek()
		if quote != 0 {
			if ch == quote {
				quote = 0
			}
			p.advance()
			continue
		}
		if ch == '\'' || ch == '"' {
			quote = ch
			p.advance()
			continue
		}
		if ch == ']' {
			content := strings.TrimSpace(p.input[start:p.pos])
			p.advance()
			return cssAttributePredicate(content)
		}
		p.advance()
	}
	return "", fmt.Errorf("unterminated attribute selector")
}

func cssAttributePredicate(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("empty attribute selector")
	}

	name, value, ok := strings.Cut(content, "*=")
	if ok {
		name = strings.TrimSpace(name)
		value = trimCSSAttributeValue(value)
		if name == "" || value == "" {
			return "", fmt.Errorf("invalid contains attribute selector")
		}
		return "contains(@" + name + ", " + xpathLiteral(value) + ")", nil
	}

	name, value, ok = strings.Cut(content, "=")
	if ok {
		name = strings.TrimSpace(name)
		value = trimCSSAttributeValue(value)
		if name == "" {
			return "", fmt.Errorf("invalid attribute selector")
		}
		return "@" + name + "=" + xpathLiteral(value), nil
	}

	name = strings.TrimSpace(content)
	if name == "" {
		return "", fmt.Errorf("empty attribute selector")
	}
	return "@" + name, nil
}

func trimCSSAttributeValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return strings.TrimSpace(value)
}

func isCSSIdentChar(ch byte) bool {
	return ch == '-' || ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9'
}

func xpathLiteral(value string) string {
	if !strings.Contains(value, "'") {
		return "'" + value + "'"
	}
	if !strings.Contains(value, `"`) {
		return `"` + value + `"`
	}

	parts := strings.Split(value, "'")
	quoted := make([]string, 0, len(parts)*2-1)
	for index, part := range parts {
		if part != "" {
			quoted = append(quoted, "'"+part+"'")
		}
		if index < len(parts)-1 {
			quoted = append(quoted, `"'"`)
		}
	}
	return "concat(" + strings.Join(quoted, ", ") + ")"
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
