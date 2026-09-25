package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageprofileexperiment"
)

type options struct{ input, output, check string }

func run(args []string, stdout, stderr io.Writer) int {
	var value options
	flags := flag.NewFlagSet("language-profile-experiment", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&value.input, "input", "", "fixed experiment input")
	flags.StringVar(&value.output, "output", "", "write report")
	flags.StringVar(&value.check, "check", "", "check report replay")
	if err := flags.Parse(args); err != nil || value.input == "" || (value.output == "") == (value.check == "") {
		fmt.Fprintln(stderr, "usage: language-profile-experiment -input <json> (-output <json> | -check <json>)")
		return 2
	}
	input, err := languageprofileexperiment.ReadInput(value.input)
	if err != nil {
		fmt.Fprintf(stderr, "language-profile-experiment: input: %v\n", err)
		return 2
	}
	report := languageprofileexperiment.Evaluate(input)
	if value.check != "" {
		expected, err := languageprofileexperiment.ReadReport(value.check)
		if err != nil || !languageprofileexperiment.Equal(report, expected) {
			fmt.Fprintln(stderr, "language-profile-experiment: replay mismatch")
			return 1
		}
	} else if err := languageprofileexperiment.WriteReport(value.output, report); err != nil {
		fmt.Fprintf(stderr, "language-profile-experiment: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "language profile: %s %d/%d\n", report.Decision,
		report.Summary.Coordinates.Satisfied, report.Summary.Coordinates.Total)
	if report.Decision == "PASS" {
		return 0
	}
	return 1
}
