package policycompilation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGoGuardPipelineCanonicalCandidateUsesFrozenNativeOracle(t *testing.T) {
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(directory, "..", "..", ".."))
	temp := t.TempDir()
	view := freezeGoGuardNativeView(t, root, temp)
	trialView, oraclePath, oracleDigest := goReturnGuardOracleView(t, view, temp)
	policyPath := filepath.Join(root, "examples", "meta-policy-compilation", "policy.gooo")
	policySource, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	program := goGuardPipelineFixture(goReturnGuardOriginal, "runMetaPolicyGenerationProfile")
	plan, err := compileGoGuardPipeline("canonical-guard.gooo", program)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := ProposeGoErrorGuardPipeline("canonical-guard.gooo", program, goReturnGuardOriginal)
	if err != nil {
		t.Fatal(err)
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, "canonical-guard.gooo"), program)
	pattern := "^(TestGoooGuardProfileOutputDelivery|TestMetaPolicyProfileWritesExactExternalFourFileBoundary)$"
	before, beforeCode := runGoGuardNativeOverlayPaths(t, root, temp, "before", goReturnGuardOriginal, trialView, plan.guard.function, pattern, goReturnGuardNativePaths)
	after, afterCode := runGoGuardNativeOverlayPaths(t, root, temp, "after", []byte(proposal.CandidateSource), trialView, plan.guard.function, pattern, goReturnGuardNativePaths)
	if beforeCode != 1 || afterCode != 0 {
		t.Fatalf("canonical native exits before=%d after=%d", beforeCode, afterCode)
	}
	for index, path := range goReturnGuardNativePaths {
		wantBefore := "pass"
		if index == 0 {
			wantBefore = "fail"
		}
		if len(before[path]) != 1 || before[path][0] != wantBefore || len(after[path]) != 1 || after[path][0] != "pass" {
			t.Fatalf("frozen canonical leaf %s: before=%v after=%v", path, before[path], after[path])
		}
	}
	view.requireUnchanged(t, root)
	requireGoReturnGuardInput(t, policyPath, policySource)
	requireGoReturnGuardInput(t, oraclePath, goReturnGuardOracle)
	logGoGuardPipelineNativeProposal(t, proposal, oracleDigest, DigestBytes(policySource))
}

func logGoGuardPipelineNativeProposal(t *testing.T, proposal GoGuardPipelineProposal, oracle, runtimePolicy string) {
	t.Helper()
	payload, err := json.Marshal(proposal)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("gooo canonical guard proposal-json: %s", payload)
	t.Logf("gooo canonical guard witness: program=%s ir=%s guard=%s rendering=%s original=%s raw=%s canonical=%s oracle=%s runtime_policy=%s paths=5 before_pass=4 before_fail=1 after_pass=5 after_fail=0 admission=UNKNOWN utility=UNKNOWN performance=UNKNOWN",
		proposal.ProgramDigest, proposal.SemanticDigest, proposal.Guard.ActivityID, proposal.Rendering.ActivityID,
		proposal.SourceDigest, proposal.Guard.CandidateDigest, proposal.CandidateDigest, oracle, runtimePolicy)
}
