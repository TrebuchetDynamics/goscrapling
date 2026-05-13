package progress

import (
	"fmt"
	"strings"
)

func ReplaceMarker(input, kind, body string) (string, error) {
	start := "<!-- PROGRESS:START kind=" + kind + " -->"
	end := "<!-- PROGRESS:END -->"
	startIndex := strings.Index(input, start)
	if startIndex < 0 {
		return "", fmt.Errorf("progress marker %q start not found", kind)
	}
	contentStart := startIndex + len(start)
	endIndex := strings.Index(input[contentStart:], end)
	if endIndex < 0 {
		return "", fmt.Errorf("progress marker %q end not found", kind)
	}
	contentEnd := contentStart + endIndex
	replacement := "\n" + strings.TrimRight(body, "\n") + "\n"
	return input[:contentStart] + replacement + input[contentEnd:], nil
}
