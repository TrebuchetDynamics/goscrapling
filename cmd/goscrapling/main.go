package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/TrebuchetDynamics/goscrapling/internal/cli"
)

func main() {
	if err := cli.Run(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, cli.ErrParse) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
