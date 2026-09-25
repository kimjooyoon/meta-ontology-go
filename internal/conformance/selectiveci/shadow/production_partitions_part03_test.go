package shadow

import (
	"testing"
)

func configureProductionPrecedence(t *testing.T, fixture *productionFixture, name string) {
	t.Helper()
	switch name {
	case "precedence-input-over-snapshot":
		fixture.reencode(t)
		fixture.files["evidence_input.json"] = []byte("{")
	case "precedence-snapshot-over-registry":
		fixture.base = buildProductionSnapshot(t, fixture.sourceBase+"// base\n", fixture.entityID, productionPrefixedDigest("different-registry"))
		fixture.planInput.Base.Files[0].BlobDigest = productionDigest("tampered-blob")
		fixture.planInput.Base.Digest = fixture.planInput.Base.ComputedDigest()
		fixture.reencode(t)
	case "precedence-registry-over-plan":
		fixture.base = buildProductionSnapshot(t, fixture.sourceBase+"// base\n", fixture.entityID, productionPrefixedDigest("different-registry"))
		fixture.planInput.CPUCapacity = 1
		fixture.encode()
	case "precedence-plan-over-proof-fail":
		fixture.planInput.CPUCapacity = 1
		fixture.reencode(t)
		fixture.proofInput.CommandReceipts[0].Digest = productionDigest("wrong-receipt")
		fixture.encode()
	case "precedence-plan-proof-over-proof-fail":
		fixture.proofInput.Snapshots.Head.Semantic = productionDigest("tampered-semantic")
		fixture.proofInput.CommandReceipts[0].Digest = productionDigest("wrong-receipt")
		fixture.encode()
	case "precedence-proof-unknown-over-lane-ineligible":
		fixture.proofInput.InferencePath.Edges = fixture.proofInput.InferencePath.Edges[:2]
		fixture.laneInput.ActiveLeaseCount = 1
		fixture.encode()
	case "precedence-lane-unknown-over-ineligible":
		fixture.laneInput.BaseSHA = ""
		fixture.laneInput.ActiveLeaseCount = 1
		fixture.encode()
	}
}
