package shadow

import (
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func (f *productionFixture) reencode(t *testing.T) {
	t.Helper()
	plan := plannersci.Plan(f.planInput)
	if len(f.proofInput.CommandReceipts) != 0 {
		f.proofInput.PlanDigest = plan.CanonicalDigest
		f.proofInput.CommandReceipts[0].PlanDigest = plan.CanonicalDigest
		f.proofInput.CommandReceipts[0].Digest = f.proofInput.CommandReceipts[0].ExpectedDigest(f.proofInput.Snapshots)
	}
	f.encode()
}
func (f *productionFixture) encode() {
	if f.files == nil {
		f.files = map[string][]byte{}
	}
	f.files["base_snapshot.json"], _ = f.base.CanonicalJSON()
	f.files["head_snapshot.json"], _ = f.head.CanonicalJSON()
	f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
	f.files["evidence_input.json"], _ = proofsci.EncodeInput(f.proofInput)
	f.files["lane_input.json"], _ = lanesci.EncodeInputJSON(f.laneInput)
}
func rebindProductionSnapshots(t *testing.T, fixture *productionFixture) {
	t.Helper()
	registryDigest := "sha256:" + fixture.planInput.Registry.Digest
	fixture.base = buildProductionSnapshot(t, fixture.sourceBase+"// base\n", fixture.entityID, registryDigest)
	fixture.head = buildProductionSnapshot(t, fixture.sourceBase+"// head\n", fixture.entityID, registryDigest)
	baseManifest := productionManifest(t, fixture.base)
	headManifest := productionManifest(t, fixture.head)
	baseBinding := semantic.SnapshotDigests{Source: rawProductionDigest(fixture.base.Digest), Semantic: baseManifest.Digest}
	headBinding := semantic.SnapshotDigests{Source: rawProductionDigest(fixture.head.Digest), Semantic: headManifest.Digest}
	plannerPath, proofPath, inferencePath, evidenceIDs := productionProofPath(t, fixture.entityID, "urn:gooo:shadow/obligation/test", fixture.commandID, baseBinding, headBinding)
	fixture.planInput.Base, fixture.planInput.Head = baseManifest, headManifest
	fixture.planInput.ProvenancePaths = []plannersci.ProvenancePath{plannerPath}
	fixture.proofInput.Snapshots = proofsci.SnapshotBinding{Base: baseBinding, Head: headBinding}
	fixture.proofInput.RegistryDigest = fixture.planInput.Registry.Digest
	fixture.proofInput.CommandReceipts[0].RegistryDigest = fixture.planInput.Registry.Digest
	fixture.laneInput.RegistryDigest = fixture.planInput.Registry.Digest
	fixture.proofInput.Paths = []proofsci.Path{proofPath}
	fixture.proofInput.InferencePath = inferencePath
	fixture.proofInput.EvidenceIDs = evidenceIDs
	fixture.proofInput.ChangedRootIDs = []semantic.ID{productionID(fixture.entityID)}
	fixture.proofInput.SelectedCommandIDs = []semantic.ID{productionID(fixture.commandID)}
}
