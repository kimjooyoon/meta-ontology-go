package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/intervention"
)

func main() {
	sourcePath := flag.String("source", "", "Gooo source value contract")
	headSHA := flag.String("head-sha", "", "exact subject commit")
	outputPath := flag.String("output", "", "intervention report output")
	check := flag.Bool("check", false, "replay and validate the fixed intervention report")
	flag.Parse()
	if *sourcePath == "" || *headSHA == "" || *outputPath == "" {
		fail("-source, -head-sha, and -output are required")
	}
	source, err := os.ReadFile(*sourcePath)
	if err != nil {
		fail(err.Error())
	}
	report, err := intervention.Build(source, *headSHA)
	if err != nil {
		fail(err.Error())
	}
	if *check {
		if err := intervention.DeterministicReplay(report, source, *headSHA); err != nil {
			fail(err.Error())
		}
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(err.Error())
	}
	raw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	if err := os.WriteFile(*outputPath, append(raw, '\n'), 0o644); err != nil {
		fail(err.Error())
	}
	fmt.Printf("invariant transformation interventions: %s semantic-expected=%d/%d semantic-operation=%d/%d nonsemantic=%d/%d writes=%d\n",
		report.Decision, report.Denominator.SemanticExpectedChange.CasesSatisfied, report.Denominator.SemanticExpectedChange.CasesTotal,
		report.Denominator.SemanticOperationChange.CasesSatisfied, report.Denominator.SemanticOperationChange.CasesTotal,
		report.Denominator.NonSemantic.CasesSatisfied, report.Denominator.NonSemantic.CasesTotal, report.RepositoryWrites)
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
