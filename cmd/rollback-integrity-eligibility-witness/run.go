package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/rollbackintegrityeligibility"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(arguments, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if options.caseID == "suite" {
		suite := rollbackintegrityeligibility.RunSuite(options.subjectSHA)
		if err := rollbackintegrityeligibility.ValidateSuite(suite, options.subjectSHA); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeOutput(options.output, suite, stdout, stderr)
	}
	input, ok := rollbackintegrityeligibility.CaseInput(options.caseID, options.subjectSHA)
	if !ok {
		fmt.Fprintf(stderr, "unknown fixed case %q\n", options.caseID)
		return 2
	}
	report := rollbackintegrityeligibility.Evaluate(input)
	if err := rollbackintegrityeligibility.Validate(report, input); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeOutput(options.output, report, stdout, stderr)
}
