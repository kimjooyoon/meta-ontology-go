package shadow

import (
	"bytes"
	"testing"

	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func productionPartition(t *testing.T, name string) productionFixture {
	t.Helper()
	fixture := newProductionFixture(t)
	switch name {
	case "positive-selective":
	case "argv-injection-looking-data", "plan-command-injection-is-data":
		fixture.planInput.Registry.Commands[0].Argv = []string{"sh", "-c", "echo SAFE; touch /tmp/gooo-shadow-must-not-run"}
		fixture.planInput.Registry.Digest = fixture.planInput.Registry.ComputedDigest()
		rebindProductionSnapshots(t, &fixture)
		fixture.reencode(t)
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
	case "plan-digest-tamper", "plan-unknown", "plan-changed-roots-mismatch", "plan-selection-union-invalid":
		fixture.planInput.CPUCapacity = 1
		fixture.reencode(t)
	case "plan-proof-source-binding-mismatch":
		fixture.proofInput.Snapshots.Base.Source = productionDigest("tampered-source")
		fixture.encode()
	case "plan-proof-semantic-binding-mismatch":
		fixture.proofInput.Snapshots.Head.Semantic = productionDigest("tampered-semantic")
		fixture.encode()
	case "plan-proof-plan-digest-mismatch":
		fixture.proofInput.PlanDigest = productionDigest("tampered-plan")
		fixture.encode()
	case "plan-proof-changed-roots-mismatch":
		fixture.proofInput.ChangedRootIDs = []semantic.ID{productionID(fixture.otherID)}
		fixture.encode()
	case "plan-proof-selected-union-mismatch":
		fixture.proofInput.SelectedCommandIDs = []semantic.ID{productionID(fixture.otherID)}
		fixture.encode()
	case "proof-fail-closed":
		fixture.proofInput.CommandReceipts[0].Digest = productionDigest("wrong-receipt")
		fixture.encode()
	case "proof-unknown":
		fixture.proofInput.InferencePath.Edges = fixture.proofInput.InferencePath.Edges[:2]
		fixture.encode()
	case "lane-unknown":
		fixture.laneInput.BaseSHA = ""
		fixture.encode()
	case "lane-ineligible":
		fixture.laneInput.ActiveLeaseCount = 1
		fixture.encode()
	case "lane-digest-tamper":
		fixture.laneInput.AheadCount = -1
		fixture.encode()
	case "malformed-unknown-field":
		fixture.encode()
		fixture.files["evidence_input.json"] = append(fixture.files["evidence_input.json"], ' ')
		fixture.files["evidence_input.json"] = bytes.Replace(fixture.files["evidence_input.json"], []byte("{"), []byte(`{"extra":true,`), 1)
	case "malformed-duplicate-key":
		fixture.encode()
		fixture.files["plan_input.json"] = bytes.Replace(fixture.files["plan_input.json"], []byte(`"schema_version":`), []byte(`"schema_version":"gooo/selective-ci/v1","schema_version":`), 1)
	case "malformed-trailing-json":
		fixture.encode()
		fixture.files["lane_input.json"] = append(fixture.files["lane_input.json"], []byte(` {}`)...)
	case "incomplete-required-field":
		fixture.encode()
		fixture.files["base_snapshot.json"] = bytes.Replace(fixture.files["base_snapshot.json"], []byte(`,"digest":"`+fixture.base.Digest+`"`), nil, 1)
	case "permutation-stable":
		fixture.planInput.Registry.Commands = append([]plannersci.Command(nil), fixture.planInput.Registry.Commands...)
		for left, right := 0, len(fixture.proofInput.InferencePath.Edges)-1; left < right; left, right = left+1, right-1 {
			fixture.proofInput.InferencePath.Edges[left], fixture.proofInput.InferencePath.Edges[right] = fixture.proofInput.InferencePath.Edges[right], fixture.proofInput.InferencePath.Edges[left]
		}
		for left, right := 0, len(fixture.proofInput.InferencePath.Evidence)-1; left < right; left, right = left+1, right-1 {
			fixture.proofInput.InferencePath.Evidence[left], fixture.proofInput.InferencePath.Evidence[right] = fixture.proofInput.InferencePath.Evidence[right], fixture.proofInput.InferencePath.Evidence[left]
		}
		fixture.reencode(t)
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
		fixture.proofInput.CommandReceipts[0].Digest = productionDigest("wrong-receipt")
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
	default:
		t.Fatalf("unmapped production partition %q", name)
	}
	return fixture
}
