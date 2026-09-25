package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/symbolicinvocationusecase"
)

func main() {
	inputPath := flag.String("input", "", "path to symbolic-reader-request-result.json")
	outputPath := flag.String("output", "", "path to write the user observation")
	expectedSubjectSHA := flag.String("expected-subject-sha", "", "exact source commit SHA")
	flag.Parse()
	if *inputPath == "" || *outputPath == "" || *expectedSubjectSHA == "" {
		fmt.Fprintln(os.Stderr, "reader-observation: --input, --output, and --expected-subject-sha are required")
		os.Exit(2)
	}

	input, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reader-observation: read input: %v\n", err)
		os.Exit(2)
	}
	report := symbolicinvocationusecase.EvaluateSymbolicReaderRequest(*expectedSubjectSHA, input)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "reader-observation: encode report: %v\n", err)
		os.Exit(2)
	}
	encoded = append(encoded, '\n')
	if err := os.WriteFile(*outputPath, encoded, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "reader-observation: write report: %v\n", err)
		os.Exit(2)
	}
	fmt.Printf("reader-observation: decision=%s resolution=%s coordinates=%d/%d writes=%d\n", report.Decision, report.Resolution, report.Coordinates.Satisfied, report.Coordinates.Total, report.Effects.RepositoryWrites)
	if report.Decision != "PASS" {
		os.Exit(1)
	}
}
