package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/selfimprovementcandidate"
)

func main() {
	contract := flag.String("contract", "examples/self-improvement/candidate.gooo", "Gooo candidate contract")
	head := flag.String("head-sha", "", "exact source commit")
	runID := flag.Int64("source-run-id", 0, "source observation workflow run")
	observation := flag.String("observation", "", "read-only observation receipt")
	output := flag.String("output", "", "candidate receipt output")
	check := flag.Bool("check", false, "require an exact proposed receipt")
	flag.Parse()
	if err := run(*contract, *head, *runID, *observation, *output, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(contract, head string, runID int64, observation, output string, check bool) error {
	raw, err := os.ReadFile(observation)
	if err != nil {
		return fmt.Errorf("read observation: %w", err)
	}
	report := selfimprovementcandidate.Evaluate(os.DirFS("."), contract, head, runID, raw)
	if err := selfimprovementcandidate.Validate(report, head, runID); err != nil {
		return err
	}
	if check && report.Decision != selfimprovementcandidate.DecisionProposed {
		return fmt.Errorf("candidate: %s / %s", report.Decision, report.Reason)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode candidate: %w", err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write candidate: %w", err)
	}
	fmt.Printf("candidate: %s %d (%d/%d -> target %d/%d)\n", report.Decision,
		report.Summary.CandidateCount, report.Summary.AchievedDelta, 1, report.Summary.TargetDelta, 1)
	return nil
}
