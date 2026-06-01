package diagnostics

import (
	"errors"
	"fmt"
)

const Usage = "usage: goscrapling {install [--force] [--json] | extract {get|post|put|delete|fetch|stealthy-fetch} <url> <output_file> [--css-selector <selector>] [-H <key: value>] [--timeout <seconds|milliseconds>] [--ai-targeted] [-p <key=value>] [--data <body>] [--json <json>] [--no-follow-redirects] [browser options] | shell -c <script> [--loglevel <level>]}"

var ErrParse = errors.New("parse error")

func ParseError(format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%w: %s\n%s", ErrParse, message, Usage)
}
