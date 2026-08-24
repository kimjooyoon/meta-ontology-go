package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/changedsurfacereceipteligibility"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(arguments []string, stdout, stderr io.Writer) int {
	options, err := parseOptions(arguments, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if options.caseID == "suite" {
		suite := changedsurfacereceipteligibility.RunSuite(options.subjectSHA)
		if err := changedsurfacereceipteligibility.ValidateSuite(suite, options.subjectSHA); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return writeOutput(options.output, suite, stdout, stderr)
	}
	input, ok := changedsurfacereceipteligibility.CaseInput(options.caseID, options.subjectSHA)
	if !ok {
		fmt.Fprintf(stderr, "unknown fixed case %q\n", options.caseID)
		return 2
	}
	report := changedsurfacereceipteligibility.Evaluate(input)
	if err := changedsurfacereceipteligibility.Validate(report, input); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return writeOutput(options.output, report, stdout, stderr)
}
