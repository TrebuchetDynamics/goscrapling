package formatting

import "strings"

func JoinOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ", ")
}

func JoinCodeOrDash(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, "`"+value+"`")
	}
	return strings.Join(out, ", ")
}

func EscapeCell(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "|", "\\|")
	if value == "" {
		return "-"
	}
	return value
}
