package main

import (
	"errors"
	"flag"
	"io"
)

type options struct {
	expectedSHA string
	outputDir   string
}

func parseOptions(arguments []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("source-authority-upstream-witness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.expectedSHA, "expected-sha", "", "exact pull request head")
	flags.StringVar(&result.outputDir, "output-dir", "", "new output directory")
	if err := flags.Parse(arguments); err != nil {
		return options{}, err
	}
	if result.expectedSHA == "" || result.outputDir == "" || flags.NArg() != 0 {
		return options{}, errors.New("--expected-sha and --output-dir are required")
	}
	return result, nil
}
