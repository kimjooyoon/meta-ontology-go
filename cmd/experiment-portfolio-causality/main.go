package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio"
)

type causalityOptions struct{ input, output, check string }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var options causalityOptions
	flags := flag.NewFlagSet("experiment-portfolio-causality", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.input, "input", "", "causality input")
	flags.StringVar(&options.output, "output", "", "write causality report")
	flags.StringVar(&options.check, "check", "", "check causality report replay")
	if err := flags.Parse(args); err != nil || options.input == "" || (options.output == "") == (options.check == "") {
		fmt.Fprintln(stderr, "usage: experiment-portfolio-causality -input <json> (-output <json> | -check <json>)")
		return 2
	}
	input, err := experimentportfolio.ReadCausalityInput(options.input)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-causality: input: %v\n", err)
		return 2
	}
	report := experimentportfolio.EvaluateCausality(input)
	if options.check != "" {
		expected, err := experimentportfolio.ReadCausalityReport(options.check)
		if err != nil || !experimentportfolio.EqualCausality(report, expected) {
			fmt.Fprintln(stderr, "experiment-portfolio-causality: replay mismatch")
			return 1
		}
	} else if err := experimentportfolio.WriteCausalityReport(options.output, report); err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-causality: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "causality audit: %s %s causal_cases=%d/%d digest_only_cases=%d hardcoded_fixture_cases=%d unknowns=%d\n", report.Decision, report.Resolution, report.Summary.CausalCases.Observed, report.Summary.CausalCases.Total, report.Summary.DigestOnlyCases, report.Summary.HardcodedFixtureCases, report.Summary.Unknowns)
	if report.Resolution == "EXACT" {
		return 0
	}
	return 1
}
