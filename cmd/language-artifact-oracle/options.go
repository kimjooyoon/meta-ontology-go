package main

import "flag"

type options struct {
	head, contract, source, unsupported, entry          string
	genuine, forged, unknown, legacy, independence, out string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("language-artifact-oracle", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "oracle contract")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.unsupported, "unsupported-source", "", "unsupported Gooo source")
	flags.StringVar(&value.entry, "entry", "", "selected activity")
	flags.StringVar(&value.genuine, "genuine", "", "genuine source artifact")
	flags.StringVar(&value.forged, "forged", "", "resealed forged artifact")
	flags.StringVar(&value.unknown, "unknown-decision", "", "unknown decision artifact")
	flags.StringVar(&value.legacy, "legacy-acceptance", "", "legacy acceptance artifact")
	flags.StringVar(&value.independence, "independence", "", "import graph evidence")
	flags.StringVar(&value.out, "out", "", "oracle report output")
	valid := flags.Parse(args) == nil && value.head != "" && value.contract != "" && value.source != "" &&
		value.unsupported != "" && value.entry != "" && value.genuine != "" && value.forged != "" &&
		value.unknown != "" && value.legacy != "" && value.independence != "" && value.out != ""
	return value, valid
}
