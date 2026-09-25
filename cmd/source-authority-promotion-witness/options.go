package main

import (
	"errors"
	"flag"
	"io"
)

type options struct {
	assurance, upstream, subjectSHA, output string
}

func parseOptions(args []string) (options, error) {
	set := flag.NewFlagSet("source-authority-promotion-witness", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	var result options
	set.StringVar(&result.assurance, "assurance", "", "language assurance report")
	set.StringVar(&result.upstream, "upstream", "", "upstream conformance report")
	set.StringVar(&result.subjectSHA, "subject-sha", "", "exact subject commit")
	set.StringVar(&result.output, "out", "", "exclusive output path")
	if err := set.Parse(args); err != nil {
		return result, err
	}
	if result.assurance == "" || result.upstream == "" || result.subjectSHA == "" || result.output == "" {
		return result, errors.New("--assurance, --upstream, --subject-sha, and --out are required")
	}
	return result, nil
}
