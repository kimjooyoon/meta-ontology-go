package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metacircularboundary"
)

func main() {
	sourcePath := flag.String("source", metacircularboundary.ExpectedSourcePath, "Gooo source to observe")
	headSHA := flag.String("head-sha", "", "exact 40-character commit SHA")
	output := flag.String("output", "", "receipt output path")
	flag.Parse()

	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fatal(err)
	}
	input := metacircularboundary.Input{Path: *sourcePath, HeadSHA: *headSHA, Source: source}
	report := metacircularboundary.Evaluate(input)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if *output == "" {
		_, _ = os.Stdout.Write(encoded)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*output, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("meta-circular boundary: %s %d/%d auth=%d exec=%d\n", report.Decision, report.Summary.CasesPassed, report.Summary.CasesTotal, report.Summary.ExplicitAuthorizations, report.Summary.AllowedExecutions)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
