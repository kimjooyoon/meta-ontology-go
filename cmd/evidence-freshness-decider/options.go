package main

import "flag"

type options struct {
	source, receipt, context, output string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("evidence-freshness-decider", flag.ContinueOnError)
	flags.StringVar(&value.receipt, "receipt", "", "evidence receipt")
	flags.StringVar(&value.source, "source", "", "raw Gooo source; omit when unavailable")
	flags.StringVar(&value.context, "context", "", "current context")
	flags.StringVar(&value.output, "output", "", "output verdict")
	return value, flags.Parse(args)
}
