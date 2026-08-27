package main

import (
	"flag"
	"fmt"
	"io"
)

type options struct {
	contract, source, checkReceipt string
	generated, replay, predecessor string
	resource, resourceOut          string
	head, output, expect           string
}

func parseOptions(args []string) (options, error) {
	var value options
	set := flag.NewFlagSet("workgraph-witness", flag.ContinueOnError)
	set.SetOutput(io.Discard)
	set.StringVar(&value.contract, "contract", "", "project contract")
	set.StringVar(&value.source, "source", "", "Gooo authority source")
	set.StringVar(&value.checkReceipt, "check-receipt", "", "Gooo check receipt")
	set.StringVar(&value.generated, "generated", "", "generated artifact")
	set.StringVar(&value.replay, "replay", "", "replayed generated artifact")
	set.StringVar(&value.predecessor, "predecessor", "", "unknown predecessor report")
	set.StringVar(&value.resource, "resource", "", "recorded resource sample")
	set.StringVar(&value.resourceOut, "resource-out", "", "write measured resource sample")
	set.StringVar(&value.head, "head", "", "exact head SHA")
	set.StringVar(&value.output, "out", "", "report path")
	set.StringVar(&value.expect, "expect", "", "expected decision")
	if err := set.Parse(args); err != nil { return value, err }
	if value.contract == "" || value.source == "" || value.checkReceipt == "" || value.head == "" || value.output == "" || value.expect == "" {
		return value, fmt.Errorf("contract, source, check-receipt, head, out, and expect are required")
	}
	if (value.generated == "") != (value.replay == "") {
		return value, fmt.Errorf("generated and replay must be supplied together")
	}
	if value.resource != "" && value.resourceOut != "" {
		return value, fmt.Errorf("resource and resource-out are mutually exclusive")
	}
	return value, nil
}
