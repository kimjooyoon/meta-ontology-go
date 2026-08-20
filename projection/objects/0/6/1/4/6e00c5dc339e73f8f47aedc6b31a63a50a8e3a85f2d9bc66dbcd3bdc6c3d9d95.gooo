package shadow

import (
	"bytes"
	"testing"
)

func productionPartition(t *testing.T, name string) productionFixture {
	t.Helper()
	fixture := newProductionFixture(t)
	switch name {
	case "positive-selective":
	case "argv-injection-looking-data", "plan-command-injection-is-data":
		configureProductionInjection(t, &fixture)
	case "snapshot-binding-manifest-mismatch", "snapshot-binding-stale-analyzer-digest", "registry-binding-mismatch":
		configureProductionSnapshot(t, &fixture, name)
	case "registry-binding-lane-mismatch":
		fixture.laneInput.RegistryDigest = productionDigest("different-lane-registry")
		fixture.encode()
	case "plan-digest-tamper", "plan-unknown", "plan-changed-roots-mismatch", "plan-selection-union-invalid":
		fixture.planInput.CPUCapacity = 1
		fixture.reencode(t)
	case "plan-proof-source-binding-mismatch", "plan-proof-semantic-binding-mismatch", "plan-proof-plan-digest-mismatch", "plan-proof-changed-roots-mismatch", "plan-proof-selected-union-mismatch", "proof-fail-closed", "proof-unknown":
		configureProductionProof(t, &fixture, name)
	case "lane-unknown", "lane-ineligible", "lane-digest-tamper":
		configureProductionLane(&fixture, name)
	case "malformed-unknown-field", "malformed-duplicate-key", "malformed-trailing-json", "incomplete-required-field":
		configureProductionMalformed(&fixture, name)
	case "permutation-stable":
		configureProductionPermutation(t, &fixture)
	case "precedence-input-over-snapshot", "precedence-snapshot-over-registry", "precedence-registry-over-plan", "precedence-plan-over-proof-fail", "precedence-plan-proof-over-proof-fail", "precedence-proof-unknown-over-lane-ineligible", "precedence-lane-unknown-over-ineligible":
		configureProductionPrecedence(t, &fixture, name)
	default:
		t.Fatalf("unmapped production partition %q", name)
	}
	return fixture
}
func configureProductionInjection(t *testing.T, fixture *productionFixture) {
	t.Helper()
	fixture.planInput.Registry.Commands[0].Argv = []string{"sh", "-c", "echo SAFE; touch /tmp/gooo-shadow-must-not-run"}
	fixture.planInput.Registry.Digest = fixture.planInput.Registry.ComputedDigest()
	rebindProductionSnapshots(t, fixture)
	fixture.reencode(t)
}
func configureProductionSnapshot(t *testing.T, fixture *productionFixture, name string) {
	t.Helper()
	switch name {
	case "snapshot-binding-manifest-mismatch":
		fixture.planInput.Base.Files[0].BlobDigest = productionDigest("tampered-blob")
		fixture.planInput.Base.Digest = fixture.planInput.Base.ComputedDigest()
		fixture.reencode(t)
	case "snapshot-binding-stale-analyzer-digest":
		fixture.reencode(t)
		fixture.files["base_snapshot.json"] = bytes.Replace(fixture.files["base_snapshot.json"], []byte(fixture.base.Digest), []byte("0"+fixture.base.Digest[1:]), 1)
	case "registry-binding-mismatch":
		fixture.base = buildProductionSnapshot(t, fixture.sourceBase+"// base\n", fixture.entityID, productionPrefixedDigest("different-registry"))
		fixture.encode()
	}
}
