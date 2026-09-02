package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/policycompilation"
)

func main() {
	policyPath := flag.String("policy", "", "raw Gooo policy source")
	casesPath := flag.String("cases", "", "canonical cases")
	outputDir := flag.String("output", "", "runner-temp output directory")
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
	if err := requireRunnerTempOutput(outputDir); err != nil {
		return err
	}
	repoRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	beforeDigest, beforeCount, err := repositorySnapshot(repoRoot)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(policyPath)
	if err != nil {
		return fmt.Errorf("read policy: %w", err)
	}
	cases, err := readCases(casesPath)
	if err != nil {
		return err
	}
	publicCLI, err := runPublicGoooCLI(repoRoot, policyPath, outputDir)
	if err != nil {
		return err
	}
	// Compile performs the producer's own raw Gooo parse/lower operation. It is
	// deliberately after the public CLI check so both boundaries are evidenced.
	policy, err := policycompilation.CompileNamed(policyPath, source)
	if err != nil {
		return fmt.Errorf("compile raw Gooo policy: %w", err)
	}
	judge := policycompilation.GenerateJudge(policy)
	judgeHash := policycompilation.DigestBytes(judge)
	cases = bindCaseDigests(cases, policy.SourceDigest, policy.SemanticDigest, judgeHash)
	artifact := policycompilation.PolicyArtifact{Schema: policycompilation.ArtifactSchema, Policy: policy, GeneratedJudgeHash: judgeHash}
	generated, independent, err := executeAll(judge, policy, cases)
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
	afterDigest, afterCount, err := repositorySnapshot(repoRoot)
	if err != nil {
		return err
	}
	writeSet := policycompilation.WriteSetObservation{
		RepositoryBeforeDigest: beforeDigest, RepositoryAfterDigest: afterDigest,
		RepositoryBeforeCount: beforeCount, RepositoryAfterCount: afterCount,
		RepositoryNetChangeObserved: beforeDigest != afterDigest,
		GeneratedRootClass:          "RUNNER_TEMP_ONLY",
		GeneratedFiles:              []string{"artifact.json", "generated-results.json", "independent-results.json", "judge.go", "policy.json", "receipt.json"},
		MutationAuthority:           0, PromotionAuthority: 0,
	}
	receipt, err := policycompilation.BuildReceipt(policy, artifact, judgeHash, cases, generated, independent, writeSet, publicCLI)
	if err != nil {
		return fmt.Errorf("build receipt: %w", err)
	}
	if err := policycompilation.VerifyReceipt(receipt, policy, artifact, judgeHash, cases); err != nil {
		return fmt.Errorf("verify receipt: %w", err)
	}
	return writeJSON(filepath.Join(outputDir, "receipt.json"), receipt)
}

func runPublicGoooCLI(repoRoot, policyPath, outputRoot string) (policycompilation.PublicCLIEvidence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	check := exec.CommandContext(ctx, "go", "run", "./cmd/gooo", "check", "--semantic", policyPath)
	check.Dir = repoRoot
	check.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := check.CombinedOutput(); err != nil {
		return policycompilation.PublicCLIEvidence{}, fmt.Errorf("public gooo check --semantic failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	cliOutput := filepath.Join(filepath.Dir(outputRoot), "public-gooo-cli")
	if err := os.MkdirAll(cliOutput, 0o750); err != nil {
		return policycompilation.PublicCLIEvidence{}, err
	}
	generate := exec.CommandContext(ctx, "go", "run", "./cmd/gooo", "generate", policyPath, "--out", cliOutput)
	generate.Dir = repoRoot
	generate.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := generate.CombinedOutput(); err != nil {
		return policycompilation.PublicCLIEvidence{}, fmt.Errorf("public gooo generate failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	entries, err := os.ReadDir(cliOutput)
	if err != nil || len(entries) == 0 {
		return policycompilation.PublicCLIEvidence{}, errors.New("public gooo generate produced no output")
	}
	return policycompilation.PublicCLIEvidence{Path: "gooo", CheckExit: 0, GenerateExit: 0, CheckObserved: true, GenerateObserved: true}, nil
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
		return nil, fmt.Errorf("canonical case denominator changed: got %d want %d", len(values), policycompilation.ExpectedCaseCount)
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value.ID == "" || seen[value.ID] || value.EvidenceClass != policycompilation.EvidenceSyntheticFixture || value.Provenance == "" {
			return nil, fmt.Errorf("case %q is not a unique synthetic fixture with provenance", value.ID)
		}
		if value.ValidatorExpectation != policycompilation.DecisionPass && value.ValidatorExpectation != policycompilation.DecisionFailClosed && value.ValidatorExpectation != policycompilation.DecisionUnknown {
			return nil, fmt.Errorf("case %q has unsupported validator expectation %q", value.ID, value.ValidatorExpectation)
		}
		seen[value.ID] = true
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
		if bound[index].ObservedIndependentDigest == "SEMANTIC_DIGEST_FROM_POLICY" || bound[index].ObservedIndependentDigest == "SOURCE_DIGEST_FROM_POLICY" {
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
		return err
	}
	temp, err := filepath.Abs(runnerTemp)
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(temp, output)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("output %q is not inside RUNNER_TEMP", output)
	}
	return nil
}

// repositorySnapshot hashes every regular repository file by relative path,
// mode, size, and content. It is an exact start/end boundary and does not rely
// on git status, which can omit ignored or content-only observations.
func repositorySnapshot(root string) (string, int, error) {
	type record struct{ path, mode, size, digest string }
	records := make([]record, 0, 128)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != root && entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		records = append(records, record{path: filepath.ToSlash(relative), mode: info.Mode().Perm().String(), size: fmt.Sprint(info.Size()), digest: contentDigest(data)})
		return nil
	})
	if err != nil {
		return "", 0, fmt.Errorf("snapshot repository: %w", err)
	}
	sort.Slice(records, func(i, j int) bool { return records[i].path < records[j].path })
	var builder strings.Builder
	for _, item := range records {
		fmt.Fprintf(&builder, "%s\t%s\t%s\t%s\n", item.path, item.mode, item.size, item.digest)
	}
	return contentDigest([]byte(builder.String())), len(records), nil
}

func contentDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
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
	if err := decoder.Decode(&trailing); err == nil {
		return value, fmt.Errorf("decode %s: trailing JSON", path)
	}
	return value, nil
}
