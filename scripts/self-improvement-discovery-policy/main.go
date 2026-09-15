package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/discoverypolicy"
)

type generationReport struct {
	Schema                       string                         `json:"schema"`
	SourcePath                   string                         `json:"source_path"`
	SourceDigest                 string                         `json:"source_digest"`
	EvaluatorDigest              string                         `json:"evaluator_digest"`
	GeneratedSourceDigest        string                         `json:"generated_source_digest"`
	GeneratedFiles               int                            `json:"generated_files"`
	GeneratedLines               int                            `json:"generated_lines"`
	PolicyActivitiesExpected     int                            `json:"policy_activities_expected"`
	PolicyActivitiesObserved     int                            `json:"policy_activities_observed"`
	Cases                        []discoverypolicy.DecisionCase `json:"cases"`
	GeneratedEvaluatorCasesBound int                            `json:"generated_evaluator_cases_bound"`
	UnboundCases                 int                            `json:"unbound_cases"`
	DuplicateCases               int                            `json:"duplicate_cases"`
	RegenerationByteMismatches   int                            `json:"regeneration_byte_mismatches"`
	RepositoryWrites             int                            `json:"repository_writes"`
	LocalBuildExecutions         int                            `json:"local_build_executions"`
	LocalTestExecutions          int                            `json:"local_test_executions"`
}

func main() {
	contract := flag.String("contract", "", "authoritative discovery policy .gooo")
	output := flag.String("output", "", "generated evaluator output")
	report := flag.String("report", "", "generation report output")
	expected := flag.String("expected", "", "checked-in evaluator for byte comparison")
	flag.Parse()
	if *contract == "" || *output == "" || *report == "" {
		exitError("usage: self-improvement-discovery-policy -contract FILE -output FILE -report FILE [-expected FILE]")
	}
	source, err := os.ReadFile(*contract)
	if err != nil {
		exitError(err.Error())
	}
	policy, generated, err := discoverypolicy.GenerateNamed(*contract, source)
	if err != nil {
		exitError(err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		exitError(err.Error())
	}
	if err := os.WriteFile(*output, generated, 0o644); err != nil {
		exitError(err.Error())
	}
	mismatches := 0
	if *expected != "" {
		expectedBytes, err := os.ReadFile(*expected)
		if err != nil {
			exitError(err.Error())
		}
		if string(expectedBytes) != string(generated) {
			mismatches = 1
		}
	}
	reportValue := generationReport{
		Schema: "gooo/public-self-observation-policy-generation/v1", SourcePath: *contract,
		SourceDigest: policy.SourceDigest, EvaluatorDigest: policy.EvaluatorDigest,
		GeneratedSourceDigest: digest(generated), GeneratedFiles: 1, GeneratedLines: lineCount(generated),
		PolicyActivitiesExpected: 1, PolicyActivitiesObserved: policy.ActivityCount, Cases: policy.Cases,
		GeneratedEvaluatorCasesBound: len(policy.Cases), UnboundCases: len(discoverypolicy.CaseIDs()) - len(policy.Cases),
		DuplicateCases: 0, RegenerationByteMismatches: mismatches, RepositoryWrites: 0,
		LocalBuildExecutions: 0, LocalTestExecutions: 0,
	}
	if reportValue.UnboundCases < 0 {
		reportValue.UnboundCases = 0
	}
	reportData, err := json.MarshalIndent(reportValue, "", "  ")
	if err != nil {
		exitError(err.Error())
	}
	if err := os.MkdirAll(filepath.Dir(*report), 0o755); err != nil {
		exitError(err.Error())
	}
	if err := os.WriteFile(*report, append(reportData, '\n'), 0o644); err != nil {
		exitError(err.Error())
	}
	if mismatches != 0 {
		exitError("generated discovery evaluator differs from checked-in evaluator")
	}
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func lineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, value := range data {
		if value == '\n' {
			count++
		}
	}
	if data[len(data)-1] == '\n' {
		count--
	}
	return count
}

func exitError(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
