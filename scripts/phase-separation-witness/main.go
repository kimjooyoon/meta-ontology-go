package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/phaseseparation"
)

func main() {
	source := flag.String("source", "examples/phase-separation-witness/main.gooo", "phase witness source")
	leaks := flag.String("leaks", "examples/phase-separation-witness/leaks.gooo", "phase leakage corpus")
	head := flag.String("head-sha", "", "exact source commit")
	output := flag.String("output", "", "witness receipt output")
	flag.Parse()
	if err := run(*source, *leaks, *head, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(sourcePath, leaksPath, headSHA, outputPath string) error {
	if outputPath == "" || headSHA == "" {
		return fmt.Errorf("head-sha and output are required")
	}
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	leaksBytes, err := os.ReadFile(leaksPath)
	if err != nil {
		return fmt.Errorf("read leakage corpus: %w", err)
	}
	report := phaseseparation.Build(sourcePath, sourceBytes, leaksPath, leaksBytes, headSHA)
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	fmt.Printf("phase separation witness: %s %d clean/%d leakage %d/%d transitions\n", report.Decision, report.Summary.CleanCasesPassed, report.Summary.LeakageCasesCaught, report.Summary.ClaimTransitionsPreserved, report.Summary.ClaimTransitionsTotal)
	return nil
}
