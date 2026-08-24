package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct{ caseID, subjectSHA, output string }

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var opts options
	set := flag.NewFlagSet("changed-surface-receipt-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	set.StringVar(&opts.caseID, "case", "", "fixed case ID or suite")
	set.StringVar(&opts.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&opts.output, "output", "", "report output")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if opts.caseID == "" || opts.subjectSHA == "" || opts.output == "" {
		return options{}, fmt.Errorf("case, subject-sha, and output are required")
	}
	return opts, nil
}
