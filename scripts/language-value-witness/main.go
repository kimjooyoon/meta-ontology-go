package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func main() {
	source := flag.String("source", "examples/language-value-witness/main.gooo", "Gooo value source")
	activity := flag.String("activity", "Increment", "activity to execute")
	head := flag.String("head-sha", "", "exact source commit")
	output := flag.String("output", "", "value witness receipt")
	check := flag.Bool("check", false, "require an exact value witness")
	flag.Parse()
	if err := run(*source, *activity, *head, *output, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, activity, head, output string, check bool) error {
	if output == "" {
		return fmt.Errorf("output is required")
	}
	report := valueexecution.Evaluate(os.DirFS("."), source, activity, head)
	if check {
		if err := valueexecution.Validate(report, head); err != nil {
			return err
		}
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode value witness: %w", err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write value witness: %w", err)
	}
	fmt.Printf("value witness: %s cases %d/%d counterexamples %d/%d improvement %d/%d -> %d/%d\n",
		report.Decision, report.Summary.ValueCasesPassed, report.Summary.ValueCasesTotal,
		report.Summary.CounterexamplesPassed, report.Summary.CounterexamplesTotal,
		report.Improvement.Before.Satisfied, report.Improvement.Before.Total,
		report.Improvement.After.Satisfied, report.Improvement.After.Total)
	return nil
}
