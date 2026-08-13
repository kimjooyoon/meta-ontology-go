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
	case "roundtrip":
		return runRoundTrip(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "query":
		return runQuery(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "inspect":
		return runInspect(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "graph":
		return runGraph(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "analyze":
		return runAnalyze(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "gooo: command %q is not implemented yet\n", args[0])
		return exitFailure
	}
}

func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|version> [args]")
}
