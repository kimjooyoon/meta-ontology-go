package main

import "flag"

type options struct {
	contract, receipt, source, output string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("ambiguity-budget-verifier", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.contract, "contract", "", "ambiguity budget contract")
	flags.StringVar(&value.receipt, "receipt", "", "producer receipt")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.output, "output", "", "judge result output")
	ok := flags.Parse(args) == nil && value.contract != "" && value.receipt != "" && value.source != "" && value.output != ""
	return value, ok
}
