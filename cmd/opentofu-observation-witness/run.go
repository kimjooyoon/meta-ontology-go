package main

import (
	"fmt"
	"os"

	observation "github.com/kimjooyoon/meta-ontology-go/internal/meta/opentofuobservation"
)

func run(args []string) int {
	opts, err := parseOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var contract observation.Contract
	if err := readStrict(opts.contract, &contract); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if opts.check != "" {
		return checkReport(opts, contract)
	}
	var input observation.Observation
	if err := readStrict(opts.input, &input); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	report, err := observation.Evaluate(contract, input)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := writeJSON(opts.output, report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if report.Decision != observation.DecisionPass || report.Resolution != observation.ResolutionExact {
		return 1
	}
	return 0
}

func checkReport(opts options, contract observation.Contract) int {
	var report observation.Report
	if err := readStrict(opts.check, &report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	var input observation.Observation
	if err := readStrict(opts.input, &input); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := observation.ValidateReport(report, input.SubjectSHA, contract.ID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
