package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	caseID, subjectSHA, observedCheckoutSHA, before, after, effectsBefore, effectsAfter, output string
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	set := flag.NewFlagSet("semantic-delta-receipt-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	var result options
	set.StringVar(&result.caseID, "case", "", "fixed case ID or suite")
	set.StringVar(&result.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&result.observedCheckoutSHA, "observed-checkout-sha", "", "observed checkout SHA")
	set.StringVar(&result.before, "before", "", "before .gooo source")
	set.StringVar(&result.after, "after", "", "after .gooo source")
	set.StringVar(&result.effectsBefore, "effects-before", "", "pre-execution workspace content snapshot")
	set.StringVar(&result.effectsAfter, "effects-after", "", "post-execution workspace content snapshot")
	set.StringVar(&result.output, "output", "", "JSON output path")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if result.subjectSHA == "" || result.output == "" {
		return options{}, fmt.Errorf("--subject-sha and --output are required")
	}
	if result.caseID == "" && (result.before == "" || result.after == "") {
		return options{}, fmt.Errorf("--case or both --before and --after are required")
	}
	if result.caseID != "" && result.caseID != "suite" && result.before != "" {
		return options{}, fmt.Errorf("--case cannot be combined with --before or --after")
	}
	return result, nil
}
