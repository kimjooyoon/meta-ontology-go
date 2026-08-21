package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"io"
	"os"
	"runtime"
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
	case "provenance":
		return runProvenance(args[1:], OSFileReader{}, SyntaxSourceParser{}, stdout, stderr)
	case "selective-ci":
		return runSelectiveCI(args[1:], OSFileReader{}, stdout, stderr)
	case "lsp":
		return runLSP(args[1:], input, stdout, stderr)
	case "version":
		return runVersion(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "gooo: command %q is not implemented yet\n", args[0])
		return exitFailure
	}
}
func printUsage(writer io.Writer) {
	fmt.Fprintln(writer, "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|provenance|selective-ci|lsp|version> [args]")
}

var analyzeDeltaToolchain = runtime.Version() + "|" + runtime.GOOS + "/" + runtime.GOARCH

type analyzeDeltaOptions struct {
	authority string
	goFiles   []string
}
type analyzeDeltaOutput struct {
	analyzer.SemanticNormalizedDelta
	AuthoritySemanticDigest string                        `json:"authority_semantic_digest"`
	ObservedSemanticDigest  string                        `json:"observed_semantic_digest"`
	SemanticEqual           bool                          `json:"semantic_equal"`
	WriteEffect             analyzer.ReconcileWriteEffect `json:"write_effect"`
}
type analyzeGeneratedRegion struct{ id string }
type analyzeMarkerAlias struct{ id, name string }
