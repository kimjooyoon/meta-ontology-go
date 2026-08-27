package main

import "flag"

type options struct {
	policy, source, head, sourcePath, cases, out, check string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-witness", flag.ContinueOnError)
	flags.StringVar(&value.policy, "policy", "", "Gooo quorum policy")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.sourcePath, "source-path", "", "source path")
	flags.StringVar(&value.cases, "cases", "", "case=id:path,path;id:path,path")
	flags.StringVar(&value.out, "out", "", "receipt or report output")
	flags.StringVar(&value.check, "check", "", "report to validate")
	return value, flags.Parse(args)
}
