package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/phaseseparation"
)

func main() {
	source := flag.String("source", "examples/phase-separation-witness/main.gooo", "clean Gooo source")
	leaks := flag.String("leaks", "examples/phase-separation-witness/leaks.gooo", "leakage Gooo source")
	unknown := flag.String("unknown", "examples/phase-separation-witness/unknown.gooo", "UNKNOWN Gooo source")
	head := flag.String("head-sha", "", "exact source commit")
	snapshot := flag.String("ci-snapshot", "", "read-only CI observation")
	output := flag.String("output", "", "witness receipt output")
	flag.Parse()
	if err := run(*source, *leaks, *unknown, *head, *snapshot, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(sourcePath, leaksPath, unknownPath, headSHA, snapshotPath, outputPath string) error {
	if outputPath == "" || headSHA == "" || snapshotPath == "" {
		return fmt.Errorf("head-sha, ci-snapshot, and output are required")
	}
	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}
	leaksBytes, err := os.ReadFile(leaksPath)
	if err != nil {
		return fmt.Errorf("read leakage source: %w", err)
	}
	unknownBytes, err := os.ReadFile(unknownPath)
	if err != nil {
		return fmt.Errorf("read UNKNOWN source: %w", err)
	}
	snapshotBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		return fmt.Errorf("read CI snapshot: %w", err)
	}
	var snapshot phaseseparation.CISnapshot
	if err := json.Unmarshal(snapshotBytes, &snapshot); err != nil {
		return fmt.Errorf("decode CI snapshot: %w", err)
	}
	report := phaseseparation.Build(sourcePath, sourceBytes, leaksPath, leaksBytes, unknownPath, unknownBytes, headSHA, snapshot)
	if report.Decision != phaseseparation.DecisionPass {
		return fmt.Errorf("phase separation witness did not pass: %s/%s", report.Decision, report.Reason)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode receipt: %w", err)
	}
	if err := os.WriteFile(outputPath, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	fmt.Printf("phase separation witness: %s source cases %d/%d clean %d/%d leakage %d/%d transitions %d/%d indicators %d/%d\n",
		report.Decision, report.Summary.SourceCasesProcessed, report.Summary.SourceCasesTotal,
		report.Summary.CleanCasesPassed, report.Summary.CleanCasesTotal,
		report.Summary.LeakageRejections, report.Summary.LeakageRejectionsTotal,
		report.Summary.ClaimTransitionsPreserved, report.Summary.ClaimTransitionsTotal,
		report.Summary.IndicatorsSatisfied, report.Summary.IndicatorsTotal)
	return nil
}
