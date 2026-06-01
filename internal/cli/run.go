package cli

import (
	"fmt"
	"io"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/extract"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/install"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/shell"
)

const usage = diagnostics.Usage

var ErrParse = diagnostics.ErrParse

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
	switch args[0] {
	case "install":
		return install.Run(stdout, args[1:])
	case "extract":
		return extract.Run(stdout, args[1:])
	case "shell":
		return shell.Run(stdout, args[1:])
	default:
		return parseError("unknown command %q", args[0])
	}
}

func parseError(format string, args ...any) error {
	return diagnostics.ParseError(format, args...)
}
