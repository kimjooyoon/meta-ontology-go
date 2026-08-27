package main

import "flag"

type options struct {
	head, contract, source, corpus, receipts, independence, out string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("counterexample-first-judge-witness", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "counterexample-first contract")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.corpus, "corpus", "", "counterexample scenario corpus")
	flags.StringVar(&value.receipts, "receipts", "", "compiler decision receipts")
	flags.StringVar(&value.independence, "independence", "", "producer dependency evidence")
	flags.StringVar(&value.out, "out", "", "independent judge report output")
	valid := flags.Parse(args) == nil && value.head != "" && value.contract != "" &&
		value.source != "" && value.corpus != "" && value.receipts != "" &&
		value.independence != "" && value.out != ""
	return value, valid
}
