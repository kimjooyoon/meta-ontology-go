package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func runStageAcceptedRevision(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: gooo stage-accepted-revision <candidate.gooo> --comparison <comparison.json> --out <directory>")
		return exitUsage
	}

	candidatePath := args[0]
	comparisonPath := ""
	outputDir := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--comparison":
			if comparisonPath != "" || i+1 >= len(args) {
				fmt.Fprintln(stderr, "stage-accepted-revision: invalid --comparison")
				return exitUsage
			}
			i++
			comparisonPath = args[i]
		case "--out":
			if outputDir != "" || i+1 >= len(args) {
				fmt.Fprintln(stderr, "stage-accepted-revision: invalid --out")
				return exitUsage
			}
			i++
			outputDir = args[i]
		default:
			fmt.Fprintf(stderr, "stage-accepted-revision: unknown argument %q\n", args[i])
			return exitUsage
		}
	}
	if comparisonPath == "" || outputDir == "" {
		fmt.Fprintln(stderr, "stage-accepted-revision: --comparison and --out are required")
		return exitUsage
	}

	candidateSource, err := reader.ReadFile(candidatePath)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: read candidate: %v\n", err)
		return exitFailure
	}
	comparisonBytes, err := reader.ReadFile(comparisonPath)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: read comparison: %v\n", err)
		return exitFailure
	}
	var comparison valueexecution.AcceptedRevisionNextRunComparison
	if err := json.Unmarshal(comparisonBytes, &comparison); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: decode comparison: %v\n", err)
		return exitFailure
	}
	stage, err := valueexecution.PrepareAcceptedRevisionNextRunStage(comparison, candidateSource)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: fail closed: %v\n", err)
		return exitFailure
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: create output: %v\n", err)
		return exitFailure
	}
	if err := os.WriteFile(filepath.Join(outputDir, "candidate.gooo"), candidateSource, 0o644); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: write candidate: %v\n", err)
		return exitFailure
	}
	comparisonCanonical, err := json.MarshalIndent(comparison, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: encode comparison: %v\n", err)
		return exitFailure
	}
	comparisonCanonical = append(comparisonCanonical, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "comparison.json"), comparisonCanonical, 0o644); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: write comparison: %v\n", err)
		return exitFailure
	}
	manifest, err := json.MarshalIndent(stage, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: encode manifest: %v\n", err)
		return exitFailure
	}
	manifest = append(manifest, '\n')
	if err := os.WriteFile(filepath.Join(outputDir, "next-run-manifest.json"), manifest, 0o644); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: write manifest: %v\n", err)
		return exitFailure
	}
	if err := json.NewEncoder(stdout).Encode(stage); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: write result: %v\n", err)
		return exitFailure
	}
	return exitOK
}
