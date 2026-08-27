package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/experimentportfolio"
)

type evaluateOptions struct{ input, output, check string }

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	var options evaluateOptions
	flags := flag.NewFlagSet("experiment-portfolio-evaluate", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.input, "input", "", "portfolio input")
	flags.StringVar(&options.output, "output", "", "write report")
	flags.StringVar(&options.check, "check", "", "check report replay")
	if err := flags.Parse(args); err != nil || options.input == "" || (options.output == "") == (options.check == "") {
		fmt.Fprintln(stderr, "usage: experiment-portfolio-evaluate -input <json> (-output <json> | -check <json>)")
		return 2
	}
	input, err := experimentportfolio.ReadInput(options.input)
	if err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-evaluate: input: %v\n", err)
		return 2
	}
	report := experimentportfolio.Evaluate(input)
	if options.check != "" {
		expected, err := experimentportfolio.ReadReport(options.check)
		if err != nil || !experimentportfolio.Equal(report, expected) {
			fmt.Fprintln(stderr, "experiment-portfolio-evaluate: replay mismatch")
			return 1
		}
	} else if err := experimentportfolio.WriteReport(options.output, report); err != nil {
		fmt.Fprintf(stderr, "experiment-portfolio-evaluate: output: %v\n", err)
		return 2
	}
	fmt.Fprintf(stdout, "experiment portfolio: %s %s candidates=%d\n", report.Decision, report.Resolution, report.Summary.Candidates)
	if report.Decision == "PORTFOLIO_PRESERVED" {
		return 0
	}
	return 1
}
