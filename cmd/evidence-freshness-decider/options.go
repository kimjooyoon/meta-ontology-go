package main

import "flag"

type options struct {
	receipt, context, output string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("evidence-freshness-decider", flag.ContinueOnError)
	flags.StringVar(&value.receipt, "receipt", "", "evidence receipt")
	flags.StringVar(&value.context, "context", "", "current context")
	flags.StringVar(&value.output, "output", "", "output verdict")
	return value, flags.Parse(args)
}
