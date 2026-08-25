package main

import (
	"fmt"
	"io"
)

func runExtensionCommand(args []string, stdout, stderr io.Writer) int {
	if args[0] == "emit" {
		return runEmit(args[1:], stdout, stderr)
	}
	fmt.Fprintf(stderr, "gooo: command %q is not implemented yet\n", args[0])
	return exitFailure
}
