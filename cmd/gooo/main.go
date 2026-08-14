package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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
	fmt.Fprintln(writer, "usage: gooo <check|generate|roundtrip|query|inspect|graph|analyze|provenance|lsp|version> [args]")
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

func parseAnalyzeDeltaArguments(args []string) (analyzeDeltaOptions, error) {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
	}
	o := analyzeDeltaOptions{authority: args[0]}
	for i := 1; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--go", "--generated-go", "--input":
			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
			}
			o.goFiles = append(o.goFiles, args[i+1])
			i++
		default:
			if strings.HasPrefix(arg, "-") {
				return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
			}
			o.goFiles = append(o.goFiles, arg)
		}
	}
	if len(o.goFiles) == 0 {
		return analyzeDeltaOptions{}, errors.New(analyzeDeltaUsage)
	}
	return o, nil
}

func runAnalyzeDelta(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	o, err := parseAnalyzeDeltaArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(2 * commandDeadline)
	authority, model, err := readAnalyzeAuthority(o.authority, reader, parser, deadline)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, o.authority, "read authority", err)
	}
	sources, err := readAnalyzeSources(o.goFiles, reader, model, authority, deadline)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "read Go input", err)
	}
	registry, err := analyzeRegistry(authority, sources)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "build symbol registry", err)
	}
	policy, err := analyzeMappingPolicy()
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "build mapping policy", err)
	}
	adapted, err := analyzer.AnalyzeAndAdaptSemantic(analyzer.SourceSemanticAdapterInput{Base: authority, Sources: sources, Registry: registry, Policy: policy, Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: analyzeDeltaToolchain})
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "adapt semantic delta", err)
	}
	if !semantic.CompareIR(adapted.IR, authority).SemanticEqual {
		return reportAnalyzeDeltaError(stderr, "", "reconcile signature facts", errors.New("deterministic signature facts disagree with DSL authority"))
	}
	payload, err := marshalAnalyzeDelta(adapted.NormalizedDelta, authority, adapted.IR)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "marshal semantic delta", err)
	}
	if err := writeInspectOutput(stdout, payload, deadline); err != nil {
		return reportAnalyzeDeltaError(stderr, "", "write semantic delta", err)
	}
	return exitOK
}
