package arguments

import (
	"strings"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
)

func NextValue(args []string, index *int, name string) (string, bool) {
	if *index+1 >= len(args) || args[*index+1] == "" {
		return "", false
	}
	*index = *index + 1
	return args[*index], true
}

func ParseHeader(value string) (string, string, error) {
	key, headerValue, ok := strings.Cut(value, ":")
	key = strings.TrimSpace(key)
	headerValue = strings.TrimSpace(headerValue)
	if !ok || key == "" {
		return "", "", diagnostics.ParseError("headers must use %q format", "Key: Value")
	}
	return key, headerValue, nil
}

func ParseKeyValue(value string, name string) (string, string, error) {
	key, paramValue, ok := strings.Cut(value, "=")
	key = strings.TrimSpace(key)
	if !ok || key == "" {
		return "", "", diagnostics.ParseError("%s must use %q format", name, "key=value")
	}
	return key, paramValue, nil
}
