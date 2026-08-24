package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct{ assurance, eligibility, subjectSHA, output string }

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var opts options
	set := flag.NewFlagSet("vertical-slice-closure-activation-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	set.StringVar(&opts.assurance, "assurance", "", "predecessor assurance capsule")
	set.StringVar(&opts.eligibility, "eligibility", "", "merged eligibility capsule")
	set.StringVar(&opts.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&opts.output, "out", "", "activation receipt output")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if opts.assurance == "" || opts.eligibility == "" || opts.subjectSHA == "" || opts.output == "" {
		return options{}, fmt.Errorf("assurance, eligibility, subject-sha, and out flags are required")
	}
	return opts, nil
}
