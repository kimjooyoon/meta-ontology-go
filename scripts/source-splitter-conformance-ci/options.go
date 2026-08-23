package main

import (
	"errors"
	"flag"
)

type options struct {
	contract    string
	evidence    string
	expectedSHA string
	output      string
	check       string
}

func parseOptions() (options, error) {
	var opts options
	flag.StringVar(&opts.contract, "contract", "", "conformance contract JSON")
	flag.StringVar(&opts.evidence, "evidence", "", "actor evidence JSON")
	flag.StringVar(&opts.expectedSHA, "expected-sha", "", "exact PR head SHA")
	flag.StringVar(&opts.output, "output", "", "artifact output path")
	flag.StringVar(&opts.check, "check", "", "artifact path to reproduce")
	flag.Parse()
	if opts.contract == "" {
		return options{}, errors.New("-contract is required")
	}
	if opts.evidence == "" {
		return options{}, errors.New("-evidence is required")
	}
	if opts.expectedSHA == "" {
		return options{}, errors.New("-expected-sha is required")
	}
	if opts.output == "" && opts.check == "" {
		return options{}, errors.New("-output or -check is required")
	}
	return opts, nil
}
