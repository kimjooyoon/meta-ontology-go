package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct{ subjectSHA, caseID, output string }

func parseOptions(arguments []string, stderr io.Writer) (options, error) {
	set := flag.NewFlagSet("vertical-slice-closure-eligibility-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	var result options
	set.StringVar(&result.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&result.caseID, "case", "suite", "suite or fixed case identifier")
	set.StringVar(&result.output, "output", "", "JSON output path")
	if err := set.Parse(arguments); err != nil {
		return options{}, err
	}
	if result.subjectSHA == "" || result.output == "" {
		return options{}, fmt.Errorf("--subject-sha and --output are required")
	}
	return result, nil
}
