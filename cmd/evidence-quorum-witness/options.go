package main

import "flag"

type options struct {
	mode, contract, source, head, sourcePath                      string
	role, originGroup, evidenceID, value                          string
	confidenceBPS                                                 int
	out, receipts, producerReceipt, unknownProducerReceipt, check string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("evidence-quorum-witness", flag.ContinueOnError)
	flags.StringVar(&value.mode, "mode", "evaluate", "emit or evaluate")
	flags.StringVar(&value.contract, "contract", "", "quorum contract")
	flags.StringVar(&value.source, "source", "", "Gooo source")
	flags.StringVar(&value.head, "head", "", "exact subject commit")
	flags.StringVar(&value.sourcePath, "source-path", "", "source path")
	flags.StringVar(&value.role, "role", "", "evidence role")
	flags.StringVar(&value.originGroup, "origin-group", "", "independence group")
	flags.StringVar(&value.evidenceID, "evidence-id", "", "evidence identity")
	flags.StringVar(&value.value, "value", "SUPPORTS", "claim value")
	flags.IntVar(&value.confidenceBPS, "confidence-bps", 0, "descriptive confidence, never aggregated")
	flags.StringVar(&value.out, "out", "", "receipt or report output")
	flags.StringVar(&value.receipts, "receipts", "", "comma-separated receipt paths")
	flags.StringVar(&value.producerReceipt, "producer-receipt", "", "Gooo producer execution receipt")
	flags.StringVar(&value.unknownProducerReceipt, "unknown-producer-receipt", "", "unknown producer execution receipt")
	flags.StringVar(&value.check, "check", "", "report to validate")
	return value, flags.Parse(args)
}
