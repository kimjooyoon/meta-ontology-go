package main

import "flag"

type options struct {
	contract, source, head, independence, writeSet, output, check, emitDir string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("evidence-freshness", flag.ContinueOnError)
	flags.StringVar(&value.contract, "contract", "", "freshness contract")
	flags.StringVar(&value.source, "source", "", "checked-in Gooo source")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.independence, "independence", "", "decider independence evidence")
	flags.StringVar(&value.writeSet, "write-set", "", "CI before/after write-set observation")
	flags.StringVar(&value.output, "output", "", "output report")
	flags.StringVar(&value.check, "check", "", "report to validate")
	flags.StringVar(&value.emitDir, "emit-dir", "", "directory for receipt and context fixtures")
	return value, flags.Parse(args)
}
