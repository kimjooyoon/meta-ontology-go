package artifactemit

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompileSymbolicValueContract(t *testing.T) {
	artifact := symbolicValueContractFixture(t)
	contract, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40))
	if err != nil {
		t.Fatalf("CompileSymbolicValueContract() error = %v", err)
	}
	if contract.Decision != "PASS" || contract.Resolution != "VALUE_CONTRACT_ONLY" {
		t.Fatalf("contract decision = %s/%s", contract.Decision, contract.Resolution)
	}
	if contract.Coordinates.Satisfied != 8 || contract.Coordinates.Total != 8 {
		t.Fatalf("contract coordinates = %d/%d", contract.Coordinates.Satisfied, contract.Coordinates.Total)
	}
	if len(contract.Rules) != 2 || contract.Rules[0].Decision != "READY" || contract.Rules[1].Decision != "FAIL_CLOSED" {
		t.Fatalf("contract rules = %#v", contract.Rules)
	}
	if contract.Default.Decision != "FAIL_CLOSED" || !strings.HasPrefix(contract.Digest, "sha256:") {
		t.Fatalf("contract default/digest = %#v %q", contract.Default, contract.Digest)
	}
	replay, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40))
	if err != nil || replay.Digest != contract.Digest {
		t.Fatalf("contract digest replay = %q, %v", replay.Digest, err)
	}
}

func TestCompileSymbolicValueContractRejectsUnknownExpectedVerdict(t *testing.T) {
	fixture := symbolicValueContractFixtureMap()
	fixtureVectors(fixture)[0]["expected"] = "UNKNOWN"
	assertSymbolicValueContractRejected(t, fixture)
}

func TestCompileSymbolicValueContractRejectsIncompleteAcceptVector(t *testing.T) {
	fixture := symbolicValueContractFixtureMap()
	fixtureVectors(fixture)[0]["instance"].(map[string]any)["activity"] = ""
	assertSymbolicValueContractRejected(t, fixture)
}

func assertSymbolicValueContractRejected(t *testing.T, fixture map[string]any) {
	t.Helper()
	artifact, err := json.Marshal(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CompileSymbolicValueContract(artifact, strings.Repeat("b", 40)); err == nil {
		t.Fatal("CompileSymbolicValueContract() accepted an invalid fixture")
	}
}
