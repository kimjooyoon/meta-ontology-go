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
	profilePackage := flag.String("profile-package", "metapolicycompilation", "expected policy package")
	profileNamespace := flag.String("profile-namespace", "metapolicycompilation", "expected policy namespace")
	flag.Parse()
	if *policyPath == "" || *casesPath == "" || *outputDir == "" {
		fmt.Fprintln(os.Stderr, "usage: meta-policy-compilation-witness -policy policy.gooo -cases cases.json -output DIR")
		os.Exit(2)
	}
	if err := produce(*policyPath, *casesPath, *outputDir, *profilePackage, *profileNamespace); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func produce(policyPath, casesPath, outputDir, profilePackage, profileNamespace string) error {
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
	policy, err := policycompilation.CompileForIdentity(policyPath, source, profilePackage, profileNamespace)
	if err != nil {
		return fmt.Errorf("compile raw Gooo policy: %w", err)
	}
	publicCLI, publicOutputDir, err := runPublicGoooCLI(repoRoot, policyPath, outputDir, profilePackage, profileNamespace)
	if err != nil {
		return err
	}
	publicArtifact, judge, _, err := readPublicGenerationArtifacts(publicOutputDir, policy)
	if err != nil {
		return err
	}
	judgeHash := publicArtifact.GeneratedJudgeHash
	cases = bindCaseDigests(cases, policy.SourceDigest, policy.SemanticDigest, judgeHash)
	artifact := publicArtifact
	generated, independent, err := executeAll(judge, policy, cases)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	if err := copyFile(filepath.Join(publicOutputDir, "policy.json"), filepath.Join(outputDir, "policy.json")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(publicOutputDir, "artifact.json"), filepath.Join(outputDir, "artifact.json")); err != nil {
		return err
	}
	if err := copyFile(filepath.Join(publicOutputDir, "judge.go"), filepath.Join(outputDir, "judge.go")); err != nil {
		return err
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

func runPublicGoooCLI(repoRoot, policyPath, outputRoot, profilePackage, profileNamespace string) (policycompilation.PublicCLIEvidence, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	check := exec.CommandContext(ctx, "go", "run", "./cmd/gooo", "check", "--semantic", policyPath)
	check.Dir = repoRoot
	check.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := check.CombinedOutput(); err != nil {
		return policycompilation.PublicCLIEvidence{}, "", fmt.Errorf("public gooo check --semantic failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	cliOutput := filepath.Clean(outputRoot) + "-public"
	if err := os.MkdirAll(cliOutput, 0o750); err != nil {
		return policycompilation.PublicCLIEvidence{}, "", err
	}
	generate := exec.CommandContext(ctx, "go", "run", "./cmd/gooo", "generate", policyPath, "--profile", policycompilation.PublicProfileID, "--profile-package", profilePackage, "--profile-namespace", profileNamespace, "--profile-project-root", repoRoot, "--out", cliOutput)
	generate.Dir = repoRoot
	generate.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := generate.CombinedOutput(); err != nil {
		return policycompilation.PublicCLIEvidence{}, "", fmt.Errorf("public gooo generate failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	entries, err := os.ReadDir(cliOutput)
	if err != nil || len(entries) == 0 {
		return policycompilation.PublicCLIEvidence{}, "", errors.New("public gooo generate produced no output")
	}
	return policycompilation.PublicCLIEvidence{Path: "gooo", CheckExit: 0, GenerateExit: 0, CheckObserved: true, GenerateObserved: true}, cliOutput, nil
}

func readPublicGenerationArtifacts(outputDir string, policy policycompilation.CompiledPolicy) (policycompilation.PolicyArtifact, []byte, policycompilation.PublicGenerationManifest, error) {
	policyBytes, err := os.ReadFile(filepath.Join(outputDir, "policy.json"))
	if err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("read public policy artifact: %w", err)
	}
	artifactBytes, err := os.ReadFile(filepath.Join(outputDir, "artifact.json"))
	if err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("read public compiled artifact: %w", err)
	}
	judge, err := os.ReadFile(filepath.Join(outputDir, "judge.go"))
	if err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("read public generated judge: %w", err)
	}
	manifestBytes, err := os.ReadFile(filepath.Join(outputDir, "generation-manifest.json"))
	if err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("read public generation manifest: %w", err)
	}
	var publicPolicy policycompilation.CompiledPolicy
	if err := json.Unmarshal(policyBytes, &publicPolicy); err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("decode public policy artifact: %w", err)
	}
	var artifact policycompilation.PolicyArtifact
	if err := json.Unmarshal(artifactBytes, &artifact); err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("decode public compiled artifact: %w", err)
	}
	var manifest policycompilation.PublicGenerationManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("decode public generation manifest: %w", err)
	}
	if publicPolicy.SourceDigest != policy.SourceDigest || publicPolicy.SemanticDigest != policy.SemanticDigest || publicPolicy.Package != policy.Package || publicPolicy.Namespace != policy.Namespace {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, errors.New("public policy artifact is not bound to source policy")
	}
	if err := policycompilation.VerifyCompiledArtifact(artifact, policy, policycompilation.DigestBytes(judge)); err != nil {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, fmt.Errorf("verify public compiled artifact: %w", err)
	}
	if manifest.Schema != policycompilation.PublicGenerationManifestSchema || manifest.Profile != policycompilation.PublicProfileID || manifest.SourceDigest != policy.SourceDigest || manifest.SemanticDigest != policy.SemanticDigest || manifest.Package != policy.Package || manifest.Namespace != policy.Namespace || manifest.PolicyBytesDigest != policycompilation.DigestBytes(policyBytes) || manifest.ArtifactBytesDigest != policycompilation.DigestBytes(artifactBytes) || manifest.GeneratedJudgeDigest != policycompilation.DigestBytes(judge) || manifest.OutputRootClass != policycompilation.PublicGenerationOutputRootClass || manifest.ExecutionObserved || manifest.CurrentConformance != policycompilation.PublicGenerationConformanceUnknown || manifest.RepositoryWrites != 0 || manifest.MutationAuthority != 0 || manifest.PromotionAuthority != 0 {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, errors.New("public generation manifest is not a pre-execution external artifact")
	}
	wantFiles := []string{"policy.json", "artifact.json", "judge.go", "generation-manifest.json"}
	if len(manifest.GeneratedFiles) != len(wantFiles) {
		return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, errors.New("public generation file denominator changed")
	}
	for index, want := range wantFiles {
		if manifest.GeneratedFiles[index] != want {
			return policycompilation.PolicyArtifact{}, nil, policycompilation.PublicGenerationManifest{}, errors.New("public generation file set is not canonical")
		}
	}
	return artifact, judge, manifest, nil
}

func copyFile(source, target string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}
	if err := os.WriteFile(target, data, 0o640); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
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
