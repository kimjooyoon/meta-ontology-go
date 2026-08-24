package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/candidateleakage"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(arguments []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(arguments, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if options.caseID == "suite" {
		suite := candidateleakage.RunSuite(options.subjectSHA)
		if err := candidateleakage.ValidateSuite(suite, options.subjectSHA); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeOutput(options.output, suite, stdout, stderr)
	}
	input, ok := candidateleakage.CaseInput(options.caseID, options.subjectSHA)
	if !ok {
		fmt.Fprintf(stderr, "unknown fixed case %q\n", options.caseID)
		return 2
	}
	report := candidateleakage.Evaluate(input)
	if err := candidateleakage.Validate(report, input); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeOutput(options.output, report, stdout, stderr)
}
