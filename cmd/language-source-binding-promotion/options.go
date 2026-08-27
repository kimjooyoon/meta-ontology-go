package main

import "flag"

type options struct {
	contract, head, policySource, policyArtifact, policyReplay string
	producer, receipt, oracle, unknownProducer, unknownOracle string
	mismatchedOracle, independence, output, check             string
}

func parseOptions(args []string) (options, error) {
	var value options
	flags := flag.NewFlagSet("language-source-binding-promotion", flag.ContinueOnError)
	flags.StringVar(&value.contract, "contract", "", "promotion contract")
	flags.StringVar(&value.head, "head", "", "exact head SHA")
	flags.StringVar(&value.policySource, "policy-source", "", "Gooo policy source")
	flags.StringVar(&value.policyArtifact, "policy-artifact", "", "generated policy artifact")
	flags.StringVar(&value.policyReplay, "policy-replay", "", "replayed generated policy artifact")
	flags.StringVar(&value.producer, "producer", "", "source execution artifact")
	flags.StringVar(&value.receipt, "receipt", "", "source execution receipt")
	flags.StringVar(&value.oracle, "oracle", "", "independent oracle report")
	flags.StringVar(&value.unknownProducer, "unknown-producer", "", "unknown producer fixture")
	flags.StringVar(&value.unknownOracle, "unknown-oracle", "", "unknown oracle fixture")
	flags.StringVar(&value.mismatchedOracle, "mismatched-oracle", "", "mismatched oracle fixture")
	flags.StringVar(&value.independence, "independence", "", "independence evidence")
	flags.StringVar(&value.output, "output", "", "output report")
	flags.StringVar(&value.check, "check", "", "report to validate")
	return value, flags.Parse(args)
}
