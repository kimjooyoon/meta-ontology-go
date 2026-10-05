package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

const bodySearchRunUsage = "usage: gooo body-search-run --source <source.gooo> --activity <name> --cases <cases.json> [--go-bin <go1.27.1>]"

type bodySearchRunResult struct {
	Schema     string               `json:"schema"`
	Generation bodycodegen.Result   `json:"generation"`
	Execution  bodyexecution.Result `json:"execution"`
}

// runBodySearch executes one Gooo-declared IR search through model selection,
// deterministic fallback, native compilation and two independent runtime runs.
func runBodySearch(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runBodySearchContext(ctx, args, stdout, stderr)
}

func runBodySearchContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := map[string]string{"--source": "", "--activity": "", "--cases": "", "--go-bin": ""}
	for i := 0; i < len(args); i += 2 {
		value, ok := flags[args[i]]
		if !ok || value != "" || i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "--") {
			fmt.Fprintln(stderr, bodySearchRunUsage)
			return exitUsage
		}
		flags[args[i]] = args[i+1]
	}
	for _, required := range []string{"--source", "--activity", "--cases"} {
		if flags[required] == "" {
			fmt.Fprintln(stderr, bodySearchRunUsage)
			return exitUsage
		}
	}
	if ctx == nil || ctx.Err() != nil {
		fmt.Fprintln(stderr, "gooo body-search-run: execution context is unavailable")
		return exitFailure
	}
	source, err := readBodyExecutionFile(flags["--source"], 128<<10)
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: source: %v\n", err)
		return exitFailure
	}
	casesBytes, err := readBodyExecutionFile(flags["--cases"], 32<<10)
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: cases: %v\n", err)
		return exitFailure
	}
	cases, err := bodyexecution.DecodeCases(casesBytes)
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: cases: %v\n", err)
		return exitFailure
	}
	var generated, diagnostics strings.Builder
	if code := runBodyCodegenContext(ctx, []string{"--json", "--activity", flags["--activity"], flags["--source"]}, OSFileReader{}, &generated, &diagnostics); code != exitOK {
		fmt.Fprintf(stderr, "gooo body-search-run: generation failed: %s", diagnostics.String())
		return code
	}
	prior, parent, err := bodyexecution.DecodeGeneration([]byte(generated.String()))
	if err != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: generation receipt: %v\n", err)
		return exitFailure
	}
	execution, runErr := bodyexecution.Execute(ctx, flags["--source"], source, pathplan.Document{}, prior, parent, cases, flags["--go-bin"])
	result := bodySearchRunResult{Schema: "gooo/body-search-run/v1", Generation: prior, Execution: execution}
	if err := json.NewEncoder(stdout).Encode(result); err != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: output: %v\n", err)
		return exitFailure
	}
	if runErr != nil {
		fmt.Fprintf(stderr, "gooo body-search-run: execution failed: %v\n", runErr)
		return exitFailure
	}
	return exitOK
}
