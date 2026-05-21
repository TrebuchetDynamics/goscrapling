package cli

import (
	"errors"
	"fmt"
	"io"
)

const usage = "usage: goscrapling extract {get|post|put|delete} <url> <output_file> [--css-selector <selector>] [-H <key: value>] [--timeout <seconds>] [-p <key=value>] [--data <body>] [--json <json>] [--no-follow-redirects]"

var ErrParse = errors.New("parse error")

func Run(stdout, stderr io.Writer, args []string) error {
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	if len(args) == 0 {
		return parseError("missing command")
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, err := fmt.Fprintln(stdout, usage)
		return err
	}
	if args[0] != "extract" {
		return parseError("unknown command %q", args[0])
	}
	return runExtract(stdout, args[1:])
}

func parseError(format string, args ...any) error {
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%w: %s\n%s", ErrParse, message, usage)
}
