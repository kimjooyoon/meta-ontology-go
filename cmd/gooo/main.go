package main

import (
	"fmt"
	"io"
	"os"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return exitUsage
	}
	switch args[0] {
	case "check":
		return runCheck(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "generate":
		return runGenerate(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "gooo: command %q is not implemented yet\n", args[0])
		return exitFailure
	}
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <check|generate> [args]")
}
