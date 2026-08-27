package main

import "flag"

type options struct {
	head, contract, source, output string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("ambiguity-budget-witness", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "ambiguity budget contract")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.output, "output", "", "receipt output")
	ok := flags.Parse(args) == nil && value.head != "" && value.contract != "" && value.source != "" && value.output != ""
	return value, ok
}
