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
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	contract, err := ambiguitybudget.DecodeContract(contractRaw)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	receipt := ambiguitybudget.Evaluate(ambiguitybudget.Input{SubjectSHA: options.head, Contract: contract, Source: source})
	if err := ambiguitybudget.WriteReceipt(options.output, receipt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := printReceipt(receipt); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if receipt.ConformanceDecision != "PASS" {
		return 1
	}
	return 0
}

func printReceipt(receipt ambiguitybudget.Receipt) error {
	fmt.Printf("ambiguity budget: conformance=%s/%s subject=%s/%s cases=%d interventions=%d denominator=%d\n",
		receipt.ConformanceDecision, receipt.ConformanceResolution, receipt.SubjectDecision, receipt.SubjectResolution,
		receipt.Summary.CasesTotal, receipt.Summary.InterventionsTotal, receipt.Summary.FixedDenominator)
	return nil
}
