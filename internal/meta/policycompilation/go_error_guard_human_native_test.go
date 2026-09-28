package policycompilation

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var goHumanGuardNativePaths = []string{
	"TestGoooHumanGuardPublicGenerate/human-denied",
	"TestGoooHumanGuardPublicGenerate/json-denied",
	"TestGoooHumanGuardPublicGenerate/human-success",
	"TestGoooHumanGuardPublicGenerate/json-success",
	"TestGoooHumanGuardConditionalDelivery/observation-denied",
	"TestGoooHumanGuardConditionalDelivery/candidate-denied",
	"TestGoooHumanGuardConditionalDelivery/all-messages-success",
	"TestGoooHumanGuardConditionalDelivery/discovery-absent-success",
}

func TestGoHumanGuardCanonicalPublicGenerationUsesFrozenNativeOracle(t *testing.T) {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "..", "..", ".."))
	temp := t.TempDir()
	view := freezeGoGuardNativeView(t, root, temp)
	trial, oraclePath, oracleDigest := goHumanGuardOracleView(t, view, temp)
	program := goHumanGuardFixture(goHumanGuardOriginal, true)
	plan, err := compileGoGuardPipeline("human-pipeline.gooo", program)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := ProposeGoErrorGuardPipeline("human-pipeline.gooo", program, goHumanGuardOriginal)
	if err != nil {
		t.Fatal(err)
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, "human-pipeline.gooo"), program)
	pattern := "^(TestGoooHumanGuardPublicGenerate|TestGoooHumanGuardConditionalDelivery)$"
	before, beforeCode := runGoGuardNativeOverlayPaths(t, root, temp, "before", goHumanGuardOriginal, trial, plan.guard.function, pattern, goHumanGuardNativePaths)
	after, afterCode := runGoGuardNativeOverlayPaths(t, root, temp, "after", []byte(proposal.CandidateSource), trial, plan.guard.function, pattern, goHumanGuardNativePaths)
	requireGoHumanGuardNativeLeaves(t, before, after, beforeCode, afterCode)
	view.requireUnchanged(t, root)
	requireGoReturnGuardInput(t, oraclePath, goHumanGuardOracle)
	payload, err := json.Marshal(proposal)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gooo human guard proposal-json: %s", payload)
	t.Logf("gooo human guard witness: program=%s ir=%s guard=%s renderer=%s original=%s raw=%s canonical=%s effective_oracle=%s guarded_calls=3 paths=8 before_pass=5 before_fail=3 after_pass=8 after_fail=0 admission=UNKNOWN adoption=UNKNOWN utility=UNKNOWN performance=UNKNOWN",
		proposal.ProgramDigest, proposal.SemanticDigest, proposal.Guard.ActivityID, proposal.Rendering.ActivityID,
		proposal.SourceDigest, proposal.Guard.CandidateDigest, proposal.CandidateDigest, oracleDigest)
}

func requireGoHumanGuardNativeLeaves(t *testing.T, before, after map[string][]string, beforeCode, afterCode int) {
	t.Helper()
	if beforeCode != 1 || afterCode != 0 {
		t.Fatalf("human guard native exits before=%d after=%d", beforeCode, afterCode)
	}
	for index, path := range goHumanGuardNativePaths {
		want := "pass"
		if index == 0 || index == 4 || index == 5 {
			want = "fail"
		}
		if len(before[path]) != 1 || before[path][0] != want ||
			len(after[path]) != 1 || after[path][0] != "pass" {
			t.Fatalf("frozen human guard leaf %s: before=%v after=%v", path, before[path], after[path])
		}
	}
}

func goHumanGuardOracleView(t *testing.T, view goGuardNativeView, temp string) (goGuardNativeView, string, string) {
	t.Helper()
	const name = "zz_gooo_human_output_guard_oracle_test.go"
	if _, exists := view.files[name]; exists {
		t.Fatal("human output oracle would replace an active file")
	}
	oraclePath := filepath.Join(temp, "human-output-oracle.go")
	writeGoGuardNativeFile(t, oraclePath, goHumanGuardOracle)
	oracle := map[string]string{name: DigestBytes(goHumanGuardOracle)}
	for name, source := range view.files {
		if strings.HasSuffix(name, "_test.go") {
			oracle[name] = DigestBytes(source)
		}
	}
	payload, err := json.Marshal(oracle)
	if err != nil {
		t.Fatal(err)
	}
	trial := view
	trial.backing = maps.Clone(view.backing)
	trial.backing[filepath.Join(view.directory, name)] = oraclePath
	t.Logf("human output oracle: additional=%s additional_digest=%s effective_files=%d effective_digest=%s",
		name, DigestBytes(goHumanGuardOracle), len(oracle), DigestBytes(payload))
	return trial, oraclePath, DigestBytes(payload)
}
