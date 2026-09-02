package artifactemit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompileSymbolicValueReachability(t *testing.T) {
	sha := strings.Repeat("c", 40)
	artifact := symbolicValueReachabilityFixture(t)
	contract := symbolicValueReachabilityContract(t, artifact, sha)
	reachability, err := CompileSymbolicValueReachability(artifact, contract, sha)
	if err != nil {
		t.Fatalf("CompileSymbolicValueReachability() error = %v", err)
	}
	if reachability.Decision != "PASS" || reachability.Resolution != "SCHEMA_VALUE_REACHABILITY_ONLY" {
		t.Fatalf("reachability decision = %s/%s", reachability.Decision, reachability.Resolution)
	}
	if reachability.Coordinates.Satisfied != 11 || reachability.Coordinates.Total != 11 {
		t.Fatalf("reachability coordinates = %d/%d", reachability.Coordinates.Satisfied, reachability.Coordinates.Total)
	}
	if reachability.Summary.ReachableRules != 1 || reachability.Summary.DefenseOnlyRules != 1 ||
		reachability.Summary.ReachableDefaults != 0 || reachability.Summary.DefenseOnlyDefaults != 1 ||
		reachability.Summary.UnknownPolicyBranches != 0 {
		t.Fatalf("reachability summary = %#v", reachability.Summary)
	}
	if !reachability.Rules[0].ReachableAfterStructuralGate || reachability.Rules[1].ReachableAfterStructuralGate ||
		reachability.Default.ReachableAfterStructuralGate {
		t.Fatalf("reachability branches = %#v %#v", reachability.Rules, reachability.Default)
	}
	replay, err := CompileSymbolicValueReachability(artifact, contract, sha)
	if err != nil || replay.Digest != reachability.Digest {
		t.Fatalf("reachability digest replay = %q, %v", replay.Digest, err)
	}
}

func TestCompileSymbolicValueReachabilityLowersUnsupportedSchema(t *testing.T) {
	sha := strings.Repeat("c", 40)
	fixture := symbolicValueReachabilityFixtureMap()
	properties := fixture["json_schema"].(map[string]any)["properties"].(map[string]any)
	properties["inputs"].(map[string]any)["items"] = true
	artifact, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	contract := symbolicValueReachabilityContract(t, artifact, sha)
	reachability, err := CompileSymbolicValueReachability(artifact, contract, sha)
	if err != nil {
		t.Fatalf("CompileSymbolicValueReachability() error = %v", err)
	}
	if reachability.Decision != "FAIL_CLOSED" || reachability.Resolution != "INVARIANT_ONLY" ||
		reachability.Summary.UnknownPolicyBranches != 3 || reachability.Default.Reachability != "UNKNOWN" {
		t.Fatalf("unsupported reachability = %#v", reachability)
	}
}

func TestCompileSymbolicValueReachabilityRejectsSourceMismatch(t *testing.T) {
	sha := strings.Repeat("c", 40)
	artifact := symbolicValueReachabilityFixture(t)
	contract := symbolicValueReachabilityContract(t, artifact, sha)
	var changed map[string]any
	if err := json.Unmarshal(artifact, &changed); err != nil {
		t.Fatal(err)
	}
	changed["digest"] = "sha256:" + strings.Repeat("d", 64)
	changedArtifact, err := json.Marshal(changed)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompileSymbolicValueReachability(changedArtifact, contract, sha); err == nil {
		t.Fatal("CompileSymbolicValueReachability() accepted a source mismatch")
	}
}
