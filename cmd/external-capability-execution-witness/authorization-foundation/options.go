package main

import (
	"flag"
	"fmt"
)

type options struct {
	subject, foundation, metadata, prior, current string
	receipt, suite, summary                       string
}

func parseOptions(args []string) (options, error) {
	value := options{}
	flags := flag.NewFlagSet("authorization-foundation", flag.ContinueOnError)
	flags.StringVar(&value.subject, "subject-sha", "", "exact CI subject SHA")
	flags.StringVar(&value.foundation, "foundation", "", "foundation contract")
	flags.StringVar(&value.metadata, "artifact-metadata", "", "GitHub artifact metadata")
	flags.StringVar(&value.prior, "prior-receipt", "", "receipt from immutable artifact")
	flags.StringVar(&value.current, "current-receipt", "", "current bootstrap receipt")
	flags.StringVar(&value.receipt, "receipt", "", "closed receipt output")
	flags.StringVar(&value.suite, "suite", "", "suite output")
	flags.StringVar(&value.summary, "summary", "", "human summary output")
	if err := flags.Parse(args); err != nil {
		return value, err
	}
	if value.subject == "" || value.foundation == "" || value.metadata == "" || value.prior == "" ||
		value.current == "" || value.receipt == "" || value.suite == "" || value.summary == "" {
		return value, fmt.Errorf("all authorization foundation options are required")
	}
	return value, nil
}
