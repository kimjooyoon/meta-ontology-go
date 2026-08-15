package verify

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func hasCouplingFailure(evidence CouplingEvidence, reason string) bool {
	want := CouplingFailureCodePrefix + reason
	for _, failure := range evidence.Failures {
		if failure.Code != want {
			continue
		}
		if reason == "source-binding-mismatch" && (failure.Domain != CouplingDomainIntegrity || failure.Owner != couplingSurface().SemanticOwnerID || failure.Retry) {
			return false
		}
		if reason == "surface-unregistered" || reason == "ambiguous-origin" || reason == "surface-not-applicable" || reason == "no-changed-sites" {
			return failure.Domain == CouplingDomainDependency && failure.Owner == CouplingOwnerUnavailable && failure.Retry
		}
		return true
	}
	return false
}

func couplingResolutionFixtures() []couplingFixtureCase {
	return []couplingFixtureCase{
		{"unregistered changed site", "NO_DELTA", CouplingDecisionUnknown, "surface-unregistered", func(in *CouplingInput) {
			in.ChangedSites[0].Path = "internal/unknown/missing.go"
			in.ChangedSites[0].CodeSymbolID = ""
		}},
		{"not applicable only changed site", "NO_DELTA", CouplingDecisionUnknown, "surface-not-applicable", func(in *CouplingInput) {
			in.Registry.Surfaces[0].Applicability = CouplingNotApplicable
			in.Envelope.RegistryDigest = in.Registry.Digest()
		}},
		{"zero changed sites", "NO_DELTA", CouplingDecisionUnknown, "no-changed-sites", func(in *CouplingInput) { in.ChangedSites = nil }},
		{"multi-invalid-ID registry", "NO_DELTA", CouplingDecisionUnknown, "registry-invalid", mutateMalformedRegistry},
	}
}

const malformedRegistryDetail = "surface ID: invalid semantic identity: identity must have a URI scheme"

func mutateMalformedRegistry(in *CouplingInput) {
	surface := &in.Registry.Surfaces[0]
	surface.SurfaceID = "bad-surface-id"
	surface.CodeSymbolID = "bad-code-symbol-id"
	surface.SemanticOwnerID = "bad-owner-id"
	surface.ScopeID = "bad-scope-id"
	surface.SourceMapID = "bad-source-map-id"
	in.Envelope.RegistryDigest = in.Registry.Digest()
}

func malformedRegistryReplayInput(t *testing.T) CouplingInput {
	t.Helper()
	in := couplingFixture(t, "NO_DELTA")
	mutateMalformedRegistry(&in)
	return in
}

func assertMalformedRegistryEvidence(t *testing.T, evidence CouplingEvidence, want []byte) []byte {
	t.Helper()
	if evidence.RawDecision != CouplingDecisionUnknown || evidence.Enforcement != CouplingEnforcementBlock {
		t.Fatalf("value = %s/%s, want UNKNOWN/BLOCK", evidence.RawDecision, evidence.Enforcement)
	}
	if len(evidence.Failures) != 1 {
		t.Fatalf("failure count = %d, want 1: %+v", len(evidence.Failures), evidence.Failures)
	}
	failure := evidence.Failures[0]
	if failure.Code != CouplingFailureCodePrefix+"registry-invalid" || failure.Domain != CouplingDomainIntegrity || failure.Owner != CouplingOwnerUnavailable || failure.Retry || failure.Detail != malformedRegistryDetail {
		t.Fatalf("failure = %+v", failure)
	}
	canonical, err := evidence.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if want != nil && string(canonical) != string(want) {
		t.Fatalf("canonical replay differs\n got: %s\nwant: %s", canonical, want)
	}
	return canonical
}

func runMalformedRegistryReplayProcess(t *testing.T, root string) []byte {
	t.Helper()
	command := exec.Command(os.Args[0], "-test.run=^TestCouplingObserverMalformedRegistryReplay$")
	command.Dir = root
	command.Env = append(os.Environ(), "COUPLING_REPLAY_HELPER=1")
	var output bytes.Buffer
	command.Stderr = &output
	err := command.Run()
	if err != nil {
		t.Fatalf("clean replay process: %v; output=%s", err, output.String())
	}
	canonical, err := hex.DecodeString(output.String())
	if err != nil {
		t.Fatalf("clean replay output: %v", err)
	}
	return canonical
}

func TestCouplingObserverMalformedRegistryReplay(t *testing.T) {
	if os.Getenv("COUPLING_REPLAY_HELPER") == "1" {
		evidence := VerifyCoupling(malformedRegistryReplayInput(t))
		canonical := assertMalformedRegistryEvidence(t, evidence, nil)
		fmt.Fprint(os.Stderr, hex.EncodeToString(canonical))
		return
	}
	input := malformedRegistryReplayInput(t)
	want := assertMalformedRegistryEvidence(t, VerifyCoupling(input), nil)
	for i := 0; i < 32; i++ {
		assertMalformedRegistryEvidence(t, VerifyCoupling(input), want)
	}
	for i := 0; i < 4; i++ {
		if got := runMalformedRegistryReplayProcess(t, t.TempDir()); string(got) != string(want) {
			t.Fatalf("clean root replay %d differs", i)
		}
	}
}
