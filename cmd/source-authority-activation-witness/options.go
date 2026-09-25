package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	assurance, upstream, eligibility string
	subjectSHA, output               string
}

func parseOptions(args []string, stderr io.Writer) (options, error) {
	var opts options
	set := flag.NewFlagSet("source-authority-activation-witness", flag.ContinueOnError)
	set.SetOutput(stderr)
	set.StringVar(&opts.assurance, "assurance", "", "predecessor assurance capsule")
	set.StringVar(&opts.upstream, "upstream", "", "predecessor upstream capsule")
	set.StringVar(&opts.eligibility, "eligibility", "", "predecessor eligibility capsule")
	set.StringVar(&opts.subjectSHA, "subject-sha", "", "exact candidate SHA")
	set.StringVar(&opts.output, "out", "", "activation receipt output")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if opts.assurance == "" || opts.upstream == "" || opts.eligibility == "" || opts.subjectSHA == "" || opts.output == "" {
		return options{}, fmt.Errorf("all evidence, subject-sha, and out flags are required")
	}
	return opts, nil
}
