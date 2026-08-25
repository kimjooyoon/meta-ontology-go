package main

import (
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/languageprofile"
)

const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

func main() { os.Exit(runWithInput(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
func run(args []string, stdout, stderr io.Writer) int {
	return runWithInput(args, os.Stdin, stdout, stderr)
}
func runWithInput(args []string, input io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return exitUsage
	}
	switch args[0] {
	case "run":
		return runSource(args[1:], OSFileReader{}, stdout, stderr)
	case "profile":
		return runProfile(args[1:], OSFileReader{}, languageprofile.RuntimeMeasurer{}, stdout, stderr)
	case "debug":
		return runDebug(args[1:], stdout, stderr)
	case "test":
		return runLanguageTest(args[1:], OSFileReader{}, stdout, stderr)
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
	case "format":
		return runFormat(args[1:], OSFileReader{}, stdout, stderr)
	case "fix":
		return runFix(args[1:], OSFileReader{}, stdout, stderr)
	case "provenance":
		return runProvenance(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "selective-ci":
		return runSelectiveCI(args[1:], OSFileReader{}, stdout, stderr)
	case "lsp":
		return runLSP(args[1:], input, stdout, stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	default:
		return runExtensionCommand(args, stdout, stderr)
	}
}
