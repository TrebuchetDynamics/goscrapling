package cli

import (
	"io"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli/diagnostics"
	"github.com/TrebuchetDynamics/goscrapling/internal/cli/router"
)

const usage = diagnostics.Usage

var ErrParse = diagnostics.ErrParse

func Run(stdout, stderr io.Writer, args []string) error {
	return router.Run(stdout, stderr, args)
}
