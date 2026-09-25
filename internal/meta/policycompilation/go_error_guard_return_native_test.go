package policycompilation

import (
	"bytes"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var goReturnGuardNativePaths = []string{
	"TestGoooGuardProfileOutputDelivery/human-denied",
	"TestGoooGuardProfileOutputDelivery/json-denied",
	"TestGoooGuardProfileOutputDelivery/human-success",
	"TestGoooGuardProfileOutputDelivery/json-success",
	"TestMetaPolicyProfileWritesExactExternalFourFileBoundary",
}

func TestGoReturnGuardActualPublicProfileUsesFrozenNativeOracle(t *testing.T) {
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
	program := goReturnGuardFixture(goReturnGuardOriginal, "runMetaPolicyGenerationProfile")
	profile, _, _, err := compileGoErrorGuard("profile-guard.gooo", program)
	if err != nil {
		t.Fatal(err)
	}
	proposal, err := ProposeGoErrorGuard("profile-guard.gooo", program, goReturnGuardOriginal)
	if err != nil {
		t.Fatal(err)
	}
	writeGoGuardNativeFile(t, filepath.Join(temp, "profile-guard.gooo"), program)
	pattern := "^(TestGoooGuardProfileOutputDelivery|TestMetaPolicyProfileWritesExactExternalFourFileBoundary)$"
	before, beforeCode := runGoGuardNativeOverlayPaths(t, root, temp, "before", goReturnGuardOriginal, trialView, profile.function, pattern, goReturnGuardNativePaths)
	after, afterCode := runGoGuardNativeOverlayPaths(t, root, temp, "after", []byte(proposal.CandidateSource), trialView, profile.function, pattern, goReturnGuardNativePaths)
	if beforeCode != 1 || afterCode != 0 {
		t.Fatalf("native profile exits before=%d after=%d", beforeCode, afterCode)
	}
	for index, path := range goReturnGuardNativePaths {
		wantBefore := "pass"
		if index == 0 {
			wantBefore = "fail"
		}
		if len(before[path]) != 1 || before[path][0] != wantBefore || len(after[path]) != 1 || after[path][0] != "pass" {
			t.Fatalf("frozen profile leaf %s: before=%v after=%v", path, before[path], after[path])
		}
	}
	view.requireUnchanged(t, root)
	requireGoReturnGuardInput(t, policyPath, policySource)
	requireGoReturnGuardInput(t, oraclePath, goReturnGuardOracle)
	t.Logf("gooo return guard witness: program=%s ir=%s activity=%s original=%s candidate=%s base_oracle=%s effective_oracle=%s oracle_scope=ACTIVE_TEST_SOURCE_SET_PLUS_FROZEN_PROFILE_ORACLE runtime_policy=%s paths=5 before_pass=4 before_fail=1 after_pass=5 after_fail=0 adoption=UNKNOWN utility=UNKNOWN performance=UNKNOWN",
		proposal.ProgramDigest, proposal.SemanticDigest, proposal.ActivityID, proposal.SourceDigest, proposal.CandidateDigest,
		view.oracleDigest, oracleDigest, DigestBytes(policySource))
}

func goReturnGuardOracleView(t *testing.T, view goGuardNativeView, temp string) (goGuardNativeView, string, string) {
	t.Helper()
	const name = "zz_gooo_guard_profile_oracle_test.go"
	if _, exists := view.files[name]; exists {
		t.Fatal("frozen profile oracle would replace a repository file")
	}
	oraclePath := filepath.Join(temp, "profile-oracle.go")
	writeGoGuardNativeFile(t, oraclePath, goReturnGuardOracle)
	oracle := map[string]string{name: DigestBytes(goReturnGuardOracle)}
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
	t.Logf("profile oracle: additional=%s additional_digest=%s effective_files=%d effective_digest=%s", name, DigestBytes(goReturnGuardOracle), len(oracle), DigestBytes(payload))
	return trial, oraclePath, DigestBytes(payload)
}

func requireGoReturnGuardInput(t *testing.T, path string, want []byte) {
	t.Helper()
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(after, want) {
		t.Fatalf("frozen profile input changed: %s error=%v", path, err)
	}
}
