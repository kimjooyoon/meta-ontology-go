package main

import (
	"flag"
	"fmt"
	"io"
)

type config struct {
	root, before, baselineReference, after, expectedSHA, output, check string
}

func parseConfig(args []string) (config, error) {
	var value config
	flags := flag.NewFlagSet("language-readiness-transition", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&value.root, "root", "", "repository root")
	flags.StringVar(&value.before, "before", "", "baseline readiness artifact")
	flags.StringVar(&value.baselineReference, "baseline-reference", "", "verified baseline reference")
	flags.StringVar(&value.after, "after", "", "current readiness artifact")
	flags.StringVar(&value.expectedSHA, "expected-sha", "", "current commit SHA")
	flags.StringVar(&value.output, "output", "", "write improvement artifact")
	flags.StringVar(&value.check, "check", "", "check improvement artifact")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if flags.NArg() != 0 {
		return config{}, fmt.Errorf("unexpected positional arguments")
	}
	if value.root == "" || value.before == "" || value.baselineReference == "" || value.after == "" ||
		value.expectedSHA == "" {
		return config{}, fmt.Errorf("root, before, baseline-reference, after, and expected-sha are required")
	}
	if (value.output == "") == (value.check == "") {
		return config{}, fmt.Errorf("exactly one of output or check is required")
	}
	return value, nil
}
