package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	mode, subject, runID, report, policySource, generated string
	foundation, envelope, receipt, suite, summary, expected string
	runAttempt                                               int
}

func parseOptions(arguments []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("external-capability-authorization", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.mode, "mode", "check", "issue or check")
	flags.StringVar(&result.subject, "subject-sha", "", "exact CI subject")
	flags.StringVar(&result.runID, "run-id", "", "exact CI run")
	flags.IntVar(&result.runAttempt, "run-attempt", 0, "exact CI run attempt")
	flags.StringVar(&result.report, "capability-report", "", "execution report")
	flags.StringVar(&result.policySource, "policy-source", "", "Gooo policy source")
	flags.StringVar(&result.generated, "policy-generated", "", "generated policy tree")
	flags.StringVar(&result.foundation, "foundation", "", "optional policy foundation")
	flags.StringVar(&result.envelope, "envelope", "", "request envelope")
	flags.StringVar(&result.receipt, "receipt", "", "authorization receipt")
	flags.StringVar(&result.suite, "suite", "", "conformance suite")
	flags.StringVar(&result.summary, "summary", "", "Markdown summary")
	flags.StringVar(&result.expected, "expect-decision", "FAIL_CLOSED", "expected decision")
	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	common := result.subject != "" && result.runID != "" && result.runAttempt > 0 &&
		result.report != "" && result.policySource != "" && result.generated != "" &&
		result.envelope != ""
	if !common {
		return options{}, fmt.Errorf("subject, invocation, report, policy, and envelope are required")
	}
	if result.mode == "check" && (result.receipt == "" || result.suite == "" || result.summary == "") {
		return options{}, fmt.Errorf("check mode requires receipt, suite, and summary")
	}
	if result.mode != "issue" && result.mode != "check" {
		return options{}, fmt.Errorf("unknown mode %q", result.mode)
	}
	return result, nil
}
