package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const (
	defaultSource          = "examples/language-value-witness/main.gooo"
	nativePlanSource       = "examples/language-runtime-binding/main.gooo"
	nativePlanEntry        = "Produce"
	nativePlanInput        = `{"value":41}`
	nativePlanSchema        = "gooo/value-execution-plan/v1"
)

type witnessReceipt struct {
	valueexecution.Report
	RuntimePlan *runtimePlanReceipt `json:"runtime_plan,omitempty"`
}

type runtimePlanReceipt struct {
	Schema              string                   `json:"schema"`
	Decision            string                   `json:"decision"`
	SourcePath          string                   `json:"source_path"`
	SourceDigest        string                   `json:"source_digest"`
	SemanticFingerprint string                   `json:"semantic_fingerprint"`
	Entry               string                   `json:"entry"`
	Execution           valueexecution.Execution `json:"execution"`
}

func main() {
	source := flag.String("source", defaultSource, "Gooo value source")
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
	var runtimePlan *runtimePlanReceipt
	if source == defaultSource {
		plan, err := executeNativePlan()
		if err != nil {
			return err
		}
		runtimePlan = &plan
	}
	encoded, err := json.MarshalIndent(witnessReceipt{Report: report, RuntimePlan: runtimePlan}, "", "  ")
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

func executeNativePlan() (runtimePlanReceipt, error) {
	input, err := os.CreateTemp("", "gooo-runtime-binding-input-*.json")
	if err != nil {
		return runtimePlanReceipt{}, fmt.Errorf("create runtime plan input: %w", err)
	}
	inputPath := input.Name()
	defer os.Remove(inputPath)
	if _, err := input.WriteString(nativePlanInput + "\n"); err != nil {
		input.Close()
		return runtimePlanReceipt{}, fmt.Errorf("write runtime plan input: %w", err)
	}
	if err := input.Close(); err != nil {
		return runtimePlanReceipt{}, fmt.Errorf("close runtime plan input: %w", err)
	}

	command := exec.Command("go", "run", "./cmd/gooo", "run", "--json", "--entry", nativePlanEntry, "--input", inputPath, nativePlanSource)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return runtimePlanReceipt{}, fmt.Errorf("execute runtime plan CLI: %w: %s", err, stderr.String())
	}
	var receipt runtimePlanReceipt
	if err := json.Unmarshal(stdout.Bytes(), &receipt); err != nil {
		return runtimePlanReceipt{}, fmt.Errorf("decode runtime plan CLI receipt: %w", err)
	}
	if err := validateNativePlan(receipt); err != nil {
		return runtimePlanReceipt{}, err
	}
	return receipt, nil
}

func validateNativePlan(receipt runtimePlanReceipt) error {
	if receipt.Schema != nativePlanSchema || receipt.Decision != "PASS" || receipt.SourcePath != nativePlanSource || receipt.Entry != nativePlanEntry {
		return fmt.Errorf("runtime plan CLI receipt identity is not exact")
	}
	if receipt.Execution.ApplyCalls != 3 || receipt.Execution.Deliveries != 2 {
		return fmt.Errorf("runtime plan CLI execution counts are not exact: applies=%d deliveries=%d", receipt.Execution.ApplyCalls, receipt.Execution.Deliveries)
	}
	for _, activity := range []string{"ConsumeA", "ConsumeB"} {
		result, ok := receipt.Execution.Results[activity]
		if !ok || result.Value != 43 {
			return fmt.Errorf("runtime plan CLI result is not exact for %s", activity)
		}
	}
	return nil
}
