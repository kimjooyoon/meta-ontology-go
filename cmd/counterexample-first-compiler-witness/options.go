package main

import "flag"

type options struct {
	head, contract, source, corpus, out string
}

func parseOptions(args []string) (options, bool) {
	flags := flag.NewFlagSet("counterexample-first-compiler-witness", flag.ContinueOnError)
	var value options
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.contract, "contract", "", "counterexample-first contract")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.corpus, "corpus", "", "counterexample scenario corpus")
	flags.StringVar(&value.out, "out", "", "decision receipt output")
	valid := flags.Parse(args) == nil && value.head != "" && value.contract != "" &&
		value.source != "" && value.corpus != "" && value.out != ""
	return value, valid
}
