package main

import (
	"fmt"

	capability "github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/externalcapabilityexecution/authorization"
)

func run(arguments []string) error {
	options, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	report, err := readJSON[capability.Report](options.report)
	if err != nil {
		return err
	}
	policy := authorization.ObservePolicy(options.policySource, options.generated)
	if options.mode == "issue" {
		envelope := authorization.Issue(options.subject, options.runID, options.runAttempt,
			report.ReportDigest, policy)
		return writeJSON(options.envelope, envelope)
	}
	envelope, err := readJSON[authorization.Envelope](options.envelope)
	if err != nil {
		return err
	}
	foundation := authorization.Foundation{}
	if options.foundation != "" {
		foundation, err = readJSON[authorization.Foundation](options.foundation)
		if err != nil {
			return err
		}
	}
	input := authorization.Input{EnvelopeAvailable: true, ReportAvailable: true,
		Envelope: envelope, Report: report, Policy: policy, Foundation: foundation,
		Invocation: authorization.Invocation{SubjectSHA: options.subject,
			RunID: options.runID, RunAttempt: options.runAttempt}}
	receipt := authorization.Evaluate(input)
	suite := authorization.RunSuite(input)
	if err := writeJSON(options.receipt, receipt); err != nil {
		return err
	}
	if err := writeJSON(options.suite, suite); err != nil {
		return err
	}
	if err := writeFile(options.summary, []byte(authorization.Markdown(receipt, suite))); err != nil {
		return err
	}
	if receipt.Decision != options.expected || suite.Passed != suite.Total {
		return fmt.Errorf("authorization conformance failed: %s, suite %d/%d",
			receipt.Decision, suite.Passed, suite.Total)
	}
	return nil
}
