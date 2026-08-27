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
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	receipt, err := os.ReadFile(options.receipt)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	source, err := os.ReadFile(options.source)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	effects, err := os.ReadFile(options.effects)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	result, err := judge.Evaluate(contract, receipt, source, effects)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if err := judge.WriteResult(options.output, result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	fmt.Printf("ambiguity budget judge: conformance=%s/%s subject=%s/%s checks=%d\n",
		result.ConformanceDecision, result.ConformanceResolution, result.SubjectDecision, result.SubjectResolution,
		len(result.Checks))
	if result.ConformanceDecision != "PASS" {
		return 1
	}
	return 0
}
