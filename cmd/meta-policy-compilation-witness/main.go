package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	policyPath := flag.String("policy", "", "Gooo policy source")
	casesPath := flag.String("cases", "", "fixed conformance cases")
	outputDir := flag.String("output", "", "producer or consumer artifact directory")
	check := flag.Bool("check", false, "verify an existing artifact directory without rewriting it")
	flag.Parse()
	if *policyPath == "" || *casesPath == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-policy-compilation-witness -policy policy.gooo -cases cases.json -output DIR [-check]")
		os.Exit(2)
	}
	var err error
	if *check {
		err = verify(*policyPath, *casesPath, *outputDir)
	} else {
		err = produce(*policyPath, *casesPath, *outputDir)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func produce(policyPath, casesPath, outputDir string) error {
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	cases, err := readCases(casesPath)
	if err != nil {
		return err
	}
	policy, err := policycompilation.Compile(source)
	if err != nil {
		return err
	}
	cases = bindCaseDigests(cases, policy.SourceDigest)
	judge := policycompilation.GenerateJudge(policy)
	judgeHash := policycompilation.DigestBytes(judge)
	artifact := policycompilation.PolicyArtifact{Schema: policycompilation.ArtifactSchema, Policy: policy, GeneratedJudgeHash: judgeHash}
	generated, independent, err := executeAll(judge, policy, cases)
	if err != nil {
		return err
	}
	receipt, err := policycompilation.BuildReceipt(policy, artifact, judgeHash, cases, generated, independent)
	if err != nil {
		return err
	}
	if err := policycompilation.VerifyReceipt(receipt, policy, artifact, judgeHash, cases); err != nil {
		return fmt.Errorf("pre-write receipt verification: %w", err)
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
	return writeJSON(filepath.Join(outputDir, "receipt.json"), receipt)
}

func verify(policyPath, casesPath, outputDir string) error {
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	cases, err := readCases(casesPath)
	if err != nil {
		return err
	}
	policy, err := policycompilation.Compile(source)
	if err != nil {
		return err
	}
	cases = bindCaseDigests(cases, policy.SourceDigest)
	artifact, err := readJSON[policycompilation.PolicyArtifact](filepath.Join(outputDir, "artifact.json"))
	if err != nil {
		return err
	}
	judge, err := os.ReadFile(filepath.Join(outputDir, "judge.go"))
	if err != nil {
		return fmt.Errorf("read generated judge: %w", err)
	}
	judgeHash := policycompilation.DigestBytes(judge)
	if err := policycompilation.VerifyCompiledArtifact(artifact, policy, judgeHash); err != nil {
		return fmt.Errorf("consumer artifact verification: %w", err)
	}
	receipt, err := readJSON[policycompilation.Receipt](filepath.Join(outputDir, "receipt.json"))
	if err != nil {
		return err
	}
	generated, err := readJSON[[]policycompilation.DecisionResult](filepath.Join(outputDir, "generated-results.json"))
	if err != nil {
		return err
	}
	independent, err := readJSON[[]policycompilation.DecisionResult](filepath.Join(outputDir, "independent-results.json"))
	if err != nil {
		return err
	}
	if len(generated) != len(cases) || len(independent) != len(cases) {
		return errors.New("stored execution results do not cover the case denominator")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for index, input := range cases {
		replayed, err := policycompilation.ExecuteGenerated(ctx, judge, input)
		if err != nil {
			return err
		}
		if !sameResult(replayed, generated[index]) {
			return fmt.Errorf("generated judge replay differs for case %q", input.ID)
		}
		expected := policycompilation.IndependentEvaluate(policy, input)
		if !sameResult(expected, independent[index]) {
			return fmt.Errorf("independent replay differs for case %q", input.ID)
		}
	}
	if err := policycompilation.VerifyReceipt(receipt, policy, artifact, judgeHash, cases); err != nil {
		return fmt.Errorf("consumer receipt verification: %w", err)
	}
	return nil
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

func sameResult(left, right policycompilation.DecisionResult) bool {
	return left.CaseID == right.CaseID && left.Decision == right.Decision && left.Stage == right.Stage && left.Step == right.Step && left.Reason == right.Reason && left.PolicyDigest == right.PolicyDigest && left.SemanticDigest == right.SemanticDigest && left.Denominator == right.Denominator
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
		if value.Expected != policycompilation.DecisionPass && value.Expected != policycompilation.DecisionFailClosed && value.Expected != policycompilation.DecisionUnknown {
			return nil, fmt.Errorf("case %q has unsupported expected decision %q", value.ID, value.Expected)
		}
	}
	return values, nil
}

func bindCaseDigests(cases []policycompilation.Case, sourceDigest string) []policycompilation.Case {
	bound := append([]policycompilation.Case(nil), cases...)
	for index := range bound {
		if bound[index].ObservedSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			bound[index].ObservedSourceDigest = sourceDigest
		}
		if bound[index].ObservedArtifactSourceDigest == "SOURCE_DIGEST_FROM_POLICY" {
			bound[index].ObservedArtifactSourceDigest = sourceDigest
		}
		if bound[index].ObservedIndependentDigest == "SOURCE_DIGEST_FROM_POLICY" {
			bound[index].ObservedIndependentDigest = sourceDigest
		}
	}
	return bound
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
	return value, nil
}
