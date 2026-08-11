package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
)

type analyzeOutput struct {
	Registrations         []analyzer.Registration         `json:"registrations"`
	Added                 []analyzer.Fact                 `json:"added"`
	Candidates            []analyzer.Candidate            `json:"candidates"`
	ImplementationDetails []analyzer.ImplementationDetail `json:"implementation_details"`
}

func runAnalyze(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: gooo analyze <file.go> [more.go ...]")
		return exitUsage
	}
	sources := make([]analyzer.SourceFile, 0, len(args))
	for _, path := range args {
		source, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(stderr, fmt.Errorf("read %s: %w", path, err))
			return exitFailure
		}
		sources = append(sources, analyzer.SourceFile{Filename: path, Source: source})
	}
	result, err := analyzer.AnalyzePackage(sources, analyzer.NewRegistry())
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	payload := analyzeOutput{
		Registrations:         result.Registrations,
		Added:                 result.Delta.Added,
		Candidates:            result.Delta.Candidates,
		ImplementationDetails: result.Delta.ImplementationDetails,
	}
	if err := json.NewEncoder(stdout).Encode(payload); err != nil {
		fmt.Fprintln(stderr, err)
		return exitFailure
	}
	return exitOK
}
