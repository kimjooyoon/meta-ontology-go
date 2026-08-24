package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	subjectSHA, parentReport, parentObservation, parentSuite string
	capabilityReport, capabilityObservation, capabilitySuite string
	report, suite string
}

func parseOptions(arguments []string, stderr io.Writer) (options, error) {
	set := flag.NewFlagSet("external-conformance-eligibility-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	var value options
	set.StringVar(&value.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&value.parentReport, "parent-report", "", "parent execution report")
	set.StringVar(&value.parentObservation, "parent-observation", "", "parent execution observation")
	set.StringVar(&value.parentSuite, "parent-suite", "", "parent execution suite")
	set.StringVar(&value.capabilityReport, "capability-report", "", "capability report")
	set.StringVar(&value.capabilityObservation, "capability-observation", "", "capability observation")
	set.StringVar(&value.capabilitySuite, "capability-suite", "", "capability suite")
	set.StringVar(&value.report, "report", "", "eligibility report output")
	set.StringVar(&value.suite, "suite", "", "eligibility suite output")
	if err := set.Parse(arguments); err != nil { return options{}, err }
	if value.subjectSHA == "" || value.parentReport == "" || value.parentObservation == "" ||
		value.parentSuite == "" || value.capabilityReport == "" || value.capabilityObservation == "" ||
		value.capabilitySuite == "" || value.report == "" || value.suite == "" {
		return options{}, fmt.Errorf("all evidence and output flags are required")
	}
	return value, nil
}
