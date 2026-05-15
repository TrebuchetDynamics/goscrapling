package customtypes

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"unicode"
)

type TextHandler string

func (t TextHandler) String() string {
	return string(t)
}

func (t TextHandler) Get(defaultValue ...string) TextHandler {
	return t
}

func (t TextHandler) GetAll() TextHandlers {
	return TextHandlers{t}
}

func (t TextHandler) Extract() TextHandlers {
	return t.GetAll()
}

func (t TextHandler) Clean() TextHandler {
	return TextHandler(normalizeSpace(string(t)))
}

func (t TextHandler) JSON() (any, error) {
	var value any
	if err := json.Unmarshal([]byte(t), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (t TextHandler) Regex(pattern string) (TextHandlers, error) {
	return regexText(string(t), pattern)
}

func (t TextHandler) RegexFirst(pattern string, defaultValue ...string) (TextHandler, error) {
	values, err := t.Regex(pattern)
	if err != nil {
		return "", err
	}
	return values.Get(defaultValue...), nil
}

type TextHandlers []TextHandler

func (h TextHandlers) Len() int {
	return len(h)
}

func (h TextHandlers) Get(defaultValue ...string) TextHandler {
	if len(h) > 0 {
		return h[0]
	}
	if len(defaultValue) > 0 {
		return TextHandler(defaultValue[0])
	}
	return ""
}

func (h TextHandlers) GetAll() TextHandlers {
	return append(TextHandlers(nil), h...)
}

func (h TextHandlers) Extract() TextHandlers {
	return h.GetAll()
}

func (h TextHandlers) Strings() []string {
	values := make([]string, 0, len(h))
	for _, value := range h {
		values = append(values, value.String())
	}
	return values
}

func (h TextHandlers) Regex(pattern string) (TextHandlers, error) {
	values := make(TextHandlers, 0)
	for _, value := range h {
		matches, err := value.Regex(pattern)
		if err != nil {
			return nil, err
		}
		values = append(values, matches...)
	}
	return values, nil
}

func (h TextHandlers) RegexFirst(pattern string, defaultValue ...string) (TextHandler, error) {
	values, err := h.Regex(pattern)
	if err != nil {
		return "", err
	}
	return values.Get(defaultValue...), nil
}

func (h TextHandlers) JSON() (any, error) {
	return h.Get().JSON()
}

func regexText(value string, pattern string) (TextHandlers, error) {
	expression, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	matches := expression.FindAllStringSubmatch(value, -1)
	values := make(TextHandlers, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			for _, capture := range match[1:] {
				values = append(values, TextHandler(html.UnescapeString(capture)))
			}
			continue
		}
		values = append(values, TextHandler(html.UnescapeString(match[0])))
	}
	return values, nil
}

type AttributesHandler struct {
	data map[string]TextHandler
}

func newAttributesHandler(attributes map[string]string) AttributesHandler {
	data := make(map[string]TextHandler, len(attributes))
	for key, value := range attributes {
		data[strings.ToLower(key)] = TextHandler(value)
	}
	return AttributesHandler{data: data}
}

func NewAttributesHandler(attributes map[string]string) AttributesHandler {
	return newAttributesHandler(attributes)
}

func (a AttributesHandler) Len() int {
	return len(a.data)
}

func (a AttributesHandler) Get(key string) (TextHandler, bool) {
	value, ok := a.data[strings.ToLower(key)]
	return value, ok
}

func (a AttributesHandler) Map() map[string]string {
	values := make(map[string]string, len(a.data))
	for key, value := range a.data {
		values[key] = value.String()
	}
	return values
}

func (a AttributesHandler) SearchValues(keyword string, partial bool) []AttributesHandler {
	matches := make([]AttributesHandler, 0)
	for key, value := range a.data {
		candidate := value.String()
		if (!partial && candidate == keyword) || (partial && strings.Contains(candidate, keyword)) {
			matches = append(matches, newAttributesHandler(map[string]string{key: candidate}))
		}
	}
	return matches
}

func (a AttributesHandler) JSON() ([]byte, error) {
	return json.Marshal(a.Map())
}

func normalizeSpace(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}
