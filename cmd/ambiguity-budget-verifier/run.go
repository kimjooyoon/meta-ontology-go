package main

import (
	"fmt"
	"os"

	judge "github.com/kimjooyoon/meta-ontology-go/internal/meta/ambiguitybudgetjudge"
)

func run(args []string) int {
	options, ok := parseOptions(args)
	if !ok {
		return 2
	}
	contract, err := os.ReadFile(options.contract)
	if err != nil {
		return 2
	}
	receipt, err := os.ReadFile(options.receipt)
	if err != nil {
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		return 2
	}
	result, err := judge.Evaluate(contract, receipt, source)
	if err != nil {
		return 2
	}
	if err := judge.WriteResult(options.output, result); err != nil {
		return 2
	}
	fmt.Printf("ambiguity budget judge: %s %s checks=%d\n", result.Decision, result.Resolution, len(result.Checks))
	if result.Decision != "PASS" {
		return 1
	}
	return 0
}
