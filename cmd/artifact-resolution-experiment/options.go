package main

import (
	"errors"
	"flag"
	"io"
)

type options struct {
	input  string
	output string
	check  string
}

func parseOptions(args []string) (options, error) {
	set := flag.NewFlagSet("artifact-resolution-experiment", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	var parsed options
	set.StringVar(&parsed.input, "input", "", "versioned experiment input")
	set.StringVar(&parsed.output, "output", "", "report output")
	set.StringVar(&parsed.check, "check", "", "expected report")
	if err := set.Parse(args); err != nil {
		return options{}, err
	}
	if parsed.input == "" || (parsed.output == "" && parsed.check == "") {
		return options{}, errors.New("-input and one of -output or -check are required")
	}
	return parsed, nil
}
