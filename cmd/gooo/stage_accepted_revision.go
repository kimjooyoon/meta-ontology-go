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
	candidatePath, comparisonPath, outputDir, ok := parseStageAcceptedRevisionArgs(args, stderr)
	if !ok {
		return exitUsage
	}
	candidateSource, comparison, stage, ok := loadStageAcceptedRevision(candidatePath, comparisonPath, reader, stderr)
	if !ok {
		return exitFailure
	}
	return writeStageAcceptedRevision(outputDir, candidateSource, comparison, stage, stdout, stderr)
}

func parseStageAcceptedRevisionArgs(args []string, stderr io.Writer) (string, string, string, bool) {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: gooo stage-accepted-revision <candidate.gooo> --comparison <comparison.json> --out <directory>")
		return "", "", "", false
	}
	candidatePath := args[0]
	comparisonPath := ""
	outputDir := ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--comparison":
			if comparisonPath != "" || i+1 >= len(args) {
				fmt.Fprintln(stderr, "stage-accepted-revision: invalid --comparison")
				return "", "", "", false
			}
			i++
			comparisonPath = args[i]
		case "--out":
			if outputDir != "" || i+1 >= len(args) {
				fmt.Fprintln(stderr, "stage-accepted-revision: invalid --out")
				return "", "", "", false
			}
			i++
			outputDir = args[i]
		default:
			fmt.Fprintf(stderr, "stage-accepted-revision: unknown argument %q\n", args[i])
			return "", "", "", false
		}
	}
	if comparisonPath == "" || outputDir == "" {
		fmt.Fprintln(stderr, "stage-accepted-revision: --comparison and --out are required")
		return "", "", "", false
	}
	return candidatePath, comparisonPath, outputDir, true
}

func loadStageAcceptedRevision(candidatePath, comparisonPath string, reader SourceReader, stderr io.Writer) ([]byte, valueexecution.AcceptedRevisionNextRunComparison, valueexecution.AcceptedRevisionNextRunStage, bool) {
	candidateSource, err := reader.ReadFile(candidatePath)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: read candidate: %v\n", err)
		return nil, valueexecution.AcceptedRevisionNextRunComparison{}, valueexecution.AcceptedRevisionNextRunStage{}, false
	}
	comparisonBytes, err := reader.ReadFile(comparisonPath)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: read comparison: %v\n", err)
		return nil, valueexecution.AcceptedRevisionNextRunComparison{}, valueexecution.AcceptedRevisionNextRunStage{}, false
	}
	var comparison valueexecution.AcceptedRevisionNextRunComparison
	if err := json.Unmarshal(comparisonBytes, &comparison); err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: decode comparison: %v\n", err)
		return nil, valueexecution.AcceptedRevisionNextRunComparison{}, valueexecution.AcceptedRevisionNextRunStage{}, false
	}
	stage, err := valueexecution.PrepareAcceptedRevisionNextRunStage(comparison, candidateSource)
	if err != nil {
		fmt.Fprintf(stderr, "stage-accepted-revision: fail closed: %v\n", err)
		return nil, valueexecution.AcceptedRevisionNextRunComparison{}, valueexecution.AcceptedRevisionNextRunStage{}, false
	}
	return candidateSource, comparison, stage, true
}

func writeStageAcceptedRevision(outputDir string, candidateSource []byte, comparison valueexecution.AcceptedRevisionNextRunComparison, stage valueexecution.AcceptedRevisionNextRunStage, stdout, stderr io.Writer) int {
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
