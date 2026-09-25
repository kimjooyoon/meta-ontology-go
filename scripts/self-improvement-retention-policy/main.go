package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/retentionpolicy"
)

type generatedCase struct {
	ID       string `json:"id"`
	Decision string `json:"decision"`
}

type report struct {
	Schema                     string          `json:"schema"`
	SourcePath                 string          `json:"source_path"`
	SourceDigest               string          `json:"source_digest"`
	EvaluatorDigest            string          `json:"evaluator_digest"`
	GeneratedSourceDigest      string          `json:"generated_source_digest"`
	GeneratedFiles             int             `json:"generated_files"`
	GeneratedLines             int             `json:"generated_lines"`
	PolicyActivitiesExpected   int             `json:"policy_activities_expected"`
	PolicyActivitiesObserved   int             `json:"policy_activities_observed"`
	GeneratedEvaluatorCases    int             `json:"generated_evaluator_cases_bound"`
	UnboundCases               int             `json:"unbound_cases"`
	DuplicateCases             int             `json:"duplicate_cases"`
	RegenerationByteMismatches int             `json:"regeneration_byte_mismatches"`
	RepositoryWrites           int             `json:"repository_writes"`
	LocalTestExecutions        int             `json:"local_test_executions"`
	Cases                      []generatedCase `json:"cases"`
}

func main() {
	contractPath := flag.String("contract", "", "authoritative retention .gooo source")
	outputPath := flag.String("output", "", "caller-owned generated evaluator path")
	reportPath := flag.String("report", "", "caller-owned generation report path")
	expectedPath := flag.String("expected", "", "optional checked-in evaluator for byte comparison")
	flag.Parse()
	if *contractPath == "" || *outputPath == "" || *reportPath == "" {
		fail(errors.New("self-improvement-retention-policy requires contract, output, and report paths"))
	}
	source, err := os.ReadFile(*contractPath)
	if err != nil {
		fail(fmt.Errorf("read retention policy: %w", err))
	}
	policy, generated, err := retentionpolicy.GenerateNamed(*contractPath, source)
	if err != nil {
		fail(fmt.Errorf("generate retention evaluator: %w", err))
	}
	mismatches := 0
	if *expectedPath != "" {
		expected, readErr := os.ReadFile(*expectedPath)
		if readErr != nil {
			fail(fmt.Errorf("read expected evaluator: %w", readErr))
		}
		if !bytes.Equal(expected, generated) {
			mismatches = 1
		}
	}
	if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
		fail(fmt.Errorf("create evaluator output directory: %w", err))
	}
	if err := os.WriteFile(*outputPath, generated, 0o644); err != nil {
		fail(fmt.Errorf("write generated evaluator: %w", err))
	}
	rows := make([]generatedCase, 0, len(policy.Cases))
	for _, row := range policy.Cases {
		rows = append(rows, generatedCase{ID: row.ID, Decision: row.Decision})
	}
	value := report{
		Schema: "gooo/semantic-retention-policy-generation/v1", SourcePath: *contractPath,
		SourceDigest: policy.SourceDigest, EvaluatorDigest: policy.EvaluatorDigest,
		GeneratedSourceDigest: cache.HashBytes(generated).String(), GeneratedFiles: 1,
		GeneratedLines: bytes.Count(generated, []byte{'\n'}), PolicyActivitiesExpected: 1,
		PolicyActivitiesObserved: policy.ActivityCount, GeneratedEvaluatorCases: len(rows),
		UnboundCases: 0, DuplicateCases: 0, RegenerationByteMismatches: mismatches,
		RepositoryWrites: 0, LocalTestExecutions: 0, Cases: rows,
	}
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fail(fmt.Errorf("encode retention policy report: %w", err))
	}
	if err := os.MkdirAll(filepath.Dir(*reportPath), 0o755); err != nil {
		fail(fmt.Errorf("create report directory: %w", err))
	}
	if err := os.WriteFile(*reportPath, append(encoded, '\n'), 0o644); err != nil {
		fail(fmt.Errorf("write retention policy report: %w", err))
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
