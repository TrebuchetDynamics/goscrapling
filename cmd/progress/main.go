package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/TrebuchetDynamics/goscrapling/cmd/progress/command"
)

func main() {
	if err := command.Run(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, command.ErrParse) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
