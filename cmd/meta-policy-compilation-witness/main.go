package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	policyPath := flag.String("policy", "", "Gooo policy source")
	casesPath := flag.String("cases", "", "fixed conformance cases")
	outputDir := flag.String("output", "", "producer or consumer artifact directory")
	flag.Parse()
	if *policyPath == "" || *casesPath == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-policy-compilation-witness -policy policy.gooo -cases cases.json -output DIR")
		os.Exit(2)
	}
	if err := produce(*policyPath, *casesPath, *outputDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func produce(policyPath, casesPath, outputDir string) error {
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	if err := requireRunnerTempOutput(outputDir); err != nil {
		return err
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	beforeDigest, beforeCount, err := observeRepositoryWriteSet(repoRoot)
	if err != nil {
		return err
	}
	cases, err := readCases(casesPath)
	if err != nil {
		return err
	}
	policy, err := policycompilation.Compile(source)
	if err != nil {
		return err
	}
	judge := policycompilation.GenerateJudge(policy)
	judgeHash := policycompilation.DigestBytes(judge)
	cases = bindCaseDigests(cases, policy.SourceDigest, policy.SemanticDigest, judgeHash)
	artifact := policycompilation.PolicyArtifact{Schema: policycompilation.ArtifactSchema, Policy: policy, GeneratedJudgeHash: judgeHash}
	generated, independent, err := executeAll(judge, policy, cases)
	if err != nil {
		return err
	}
	receipt, err := policycompilation.BuildReceipt(policy, artifact, judgeHash, cases, generated, independent, policycompilation.WriteSetObservation{RepositoryBeforeDigest: beforeDigest, RepositoryBeforeCount: beforeCount, GeneratedRootClass: "RUNNER_TEMP_ONLY", GeneratedFiles: []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"}})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	if err := writeJSON(filepath.Join(outputDir, "policy.json"), policy); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outputDir, "artifact.json"), artifact); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "judge.go"), judge, 0o640); err != nil {
		return fmt.Errorf("write generated judge: %w", err)
	}
	if err := writeJSON(filepath.Join(outputDir, "generated-results.json"), generated); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(outputDir, "independent-results.json"), independent); err != nil {
		return err
	}
	afterDigest, afterCount, err := observeRepositoryWriteSet(repoRoot)
	if err != nil {
		return err
	}
	receipt.WriteSet = policycompilation.WriteSetObservation{
		RepositoryBeforeDigest: beforeDigest, RepositoryAfterDigest: afterDigest,
		RepositoryBeforeCount: beforeCount, RepositoryAfterCount: afterCount,
		RepositoryNetChangeObserved: beforeDigest != afterDigest,
		GeneratedRootClass:          "RUNNER_TEMP_ONLY",
		GeneratedFiles:              []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"},
		MutationAuthority:           0, PromotionAuthority: 0,
	}
	if err := policycompilation.FinalizeReceipt(&receipt); err != nil {
		return err
	}
	if err := policycompilation.VerifyReceipt(receipt, policy, artifact, judgeHash, cases); err != nil {
		return fmt.Errorf("pre-write receipt verification: %w", err)
	}
	return writeJSON(filepath.Join(outputDir, "receipt.json"), receipt)
}

func executeAll(judge []byte, policy policycompilation.CompiledPolicy, cases []policycompilation.Case) ([]policycompilation.DecisionResult, []policycompilation.DecisionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	generated := make([]policycompilation.DecisionResult, 0, len(cases))
	independent := make([]policycompilation.DecisionResult, 0, len(cases))
	for _, input := range cases {
		result, err := policycompilation.ExecuteGenerated(ctx, judge, input)
		if err != nil {
			return nil, nil, err
		}
		generated = append(generated, result)
		independent = append(independent, policycompilation.IndependentEvaluate(policy, input))
	}
	return generated, independent, nil
}

func readCases(path string) ([]policycompilation.Case, error) {
	values, err := readJSON[[]policycompilation.Case](path)
	if err != nil {
		return nil, err
	}
	if len(values) != policycompilation.ExpectedCaseCount {
		return nil, fmt.Errorf("case denominator changed: got %d want %d", len(values), policycompilation.ExpectedCaseCount)
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value.ID == "" || seen[value.ID] {
			return nil, errors.New("case IDs must be non-empty and unique")
		}
		seen[value.ID] = true
		if value.ValidatorExpectation != policycompilation.DecisionPass && value.ValidatorExpectation != policycompilation.DecisionFailClosed && value.ValidatorExpectation != policycompilation.DecisionUnknown {
			return nil, fmt.Errorf("case %q has unsupported validator expectation %q", value.ID, value.ValidatorExpectation)
		}
		if value.EvidenceClass != policycompilation.EvidenceSyntheticFixture || value.Provenance == "" {
			return nil, fmt.Errorf("case %q must declare synthetic-fixture evidence and provenance", value.ID)
		}
	}
	return values, nil
}

func bindCaseDigests(cases []policycompilation.Case, sourceDigest, semanticDigest, judgeHash string) []policycompilation.Case {
	bound := append([]policycompilation.Case(nil), cases...)
	for index := range bound {
		if bound[index].ObservedSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			bound[index].ObservedSourceDigest = sourceDigest
		}
		if bound[index].ObservedArtifactSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			bound[index].ObservedArtifactSourceDigest = sourceDigest
		}
		if bound[index].ObservedGeneratedJudgeDigest == "GENERATED_JUDGE_DIGEST_FROM_ARTIFACT" {
			bound[index].ObservedGeneratedJudgeDigest = judgeHash
		}
		if bound[index].ObservedIndependentDigest == "SOURCE_DIGEST_FROM_POLICY" || bound[index].ObservedIndependentDigest == "SEMANTIC_DIGEST_FROM_POLICY" {
			bound[index].ObservedIndependentDigest = semanticDigest
		}
	}
	return bound
}

func requireRunnerTempOutput(path string) error {
	runnerTemp := os.Getenv("RUNNER_TEMP")
	if runnerTemp == "" {
		return errors.New("RUNNER_TEMP is required: generated files must stay outside the repository")
	}
	output, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	temp, err := filepath.Abs(runnerTemp)
	if err != nil {
		return fmt.Errorf("resolve runner temp: %w", err)
	}
	relative, err := filepath.Rel(temp, output)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("output %q is not inside RUNNER_TEMP", output)
	}
	return nil
}

func observeRepositoryWriteSet(repoRoot string) (string, int, error) {
	command := exec.Command("git", "-C", repoRoot, "status", "--porcelain=v1", "--untracked-files=all")
	output, err := command.Output()
	if err != nil {
		return "", 0, fmt.Errorf("observe repository write set: %w", err)
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return policycompilation.DigestBytes(output), 0, nil
	}
	return policycompilation.DigestBytes(output), len(strings.Split(trimmed, "\n")), nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func readJSON[T any](path string) (T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("read %s: %w", path, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var value T
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode %s: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return value, fmt.Errorf("decode %s: trailing JSON", path)
	}
	return value, nil
}
