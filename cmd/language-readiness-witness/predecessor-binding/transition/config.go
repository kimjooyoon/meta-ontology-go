package main

import (
	"flag"
	"fmt"
	"io"
)

type config struct {
	root, before, after, expectedSHA, output, check string
}

func parseConfig(args []string) (config, error) {
	value := config{}
	flags := flag.NewFlagSet("predecessor-binding-transition", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&value.root, "root", "", "repository root")
	flags.StringVar(&value.before, "before", "", "predecessor binding report")
	flags.StringVar(&value.after, "after", "", "current binding report")
	flags.StringVar(&value.expectedSHA, "expected-sha", "", "current exact head SHA")
	flags.StringVar(&value.output, "output", "", "write transition")
	flags.StringVar(&value.check, "check", "", "check transition")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if value.root == "" || value.before == "" || value.after == "" || value.expectedSHA == "" {
		return config{}, fmt.Errorf("root, before, after, and expected-sha are required")
	}
	if (value.output == "") == (value.check == "") {
		return config{}, fmt.Errorf("exactly one of output or check is required")
	}
	return value, nil
}
