package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	contract string
	input    string
	output   string
	check    string
}

func parseOptions(args []string) (options, error) {
	var result options
	flags := flag.NewFlagSet("opentofu-observation-witness", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&result.contract, "contract", "", "contract JSON path")
	flags.StringVar(&result.input, "input", "", "observation JSON path")
	flags.StringVar(&result.output, "output", "", "report JSON path")
	flags.StringVar(&result.check, "check", "", "report JSON path to validate")
	if err := flags.Parse(args); err != nil {
		return options{}, err
	}
	if result.contract == "" || result.input == "" || result.output == "" {
		return options{}, fmt.Errorf("contract, input, and output are required")
	}
	return result, nil
}
