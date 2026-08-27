package main

import (
	"fmt"
	"os"

	ambiguitybudget "github.com/kimjooyoon/meta-ontology-go/internal/meta/ambiguitybudget"
)

func run(args []string) int {
	options, ok := parseOptions(args)
	if !ok {
		return 2
	}
	contractRaw, err := os.ReadFile(options.contract)
	if err != nil {
		return 2
	}
	contract, err := ambiguitybudget.DecodeContract(contractRaw)
	if err != nil {
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return 2
	}
	receipt := ambiguitybudget.Evaluate(ambiguitybudget.Input{SubjectSHA: options.head, Contract: contract, Source: source})
	if err := ambiguitybudget.WriteReceipt(options.output, receipt); err != nil {
		return 2
	}
	if err := printReceipt(receipt); err != nil {
		return 2
	}
	if receipt.Decision != "PASS" {
		return 1
	}
	return 0
}

func printReceipt(receipt ambiguitybudget.Receipt) error {
	fmt.Printf("ambiguity budget: %s %s cases=%d/%d coordinates=%d/%d\n", receipt.Decision, receipt.Resolution,
		receipt.Summary.CasesSatisfied, receipt.Summary.CasesTotal,
		receipt.Summary.CoordinatesSatisfied, receipt.Summary.CoordinatesTotal)
	return nil
}
