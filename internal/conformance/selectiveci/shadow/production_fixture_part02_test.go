package shadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func newProductionFixture(t *testing.T) productionFixture {
	t.Helper()
	entityID := "urn:gooo:shadow/entity/order"
	otherID := "urn:gooo:shadow/entity/other"
	commandID := "urn:gooo:shadow/command/test"
	obligationID := "urn:gooo:shadow/obligation/test"
	sourceBase := `package fixture

//gooo:bind id="urn:gooo:shadow/entity/order" role="HANDWRITTEN_IMPL"
func Order() {}
`
	registry := plannersci.Registry{
		SchemaVersion:   plannersci.RegistrySchemaVersion,
		PolicyDigest:    productionDigest("policy"),
		Nodes:           []impactgraph.Node{{ID: entityID, Kind: impactgraph.NodeKindSemantic}},
		DependencyEdges: []plannersci.DependencyEdge{},
		Obligations:     []plannersci.ObligationBinding{{ID: obligationID, Subject: entityID, CommandIDs: []string{commandID}}},
		Commands: []plannersci.Command{{ID: commandID, Argv: []string{"gooo-shadow-sentinel", "never-run"},
			WorkingDir: ".", CPUWorkUnits: 100, MemoryBytes: 1000}},
		GlobalGuardCommands: []plannersci.Command{},
	}
	registry.Digest = registry.ComputedDigest()
	base := buildProductionSnapshot(t, sourceBase+"// base\n", entityID, "sha256:"+registry.Digest)
	head := buildProductionSnapshot(t, sourceBase+"// head\n", entityID, "sha256:"+registry.Digest)
	baseManifest := productionManifest(t, base)
	headManifest := productionManifest(t, head)
	baseBinding := semantic.SnapshotDigests{Source: rawProductionDigest(base.Digest), Semantic: baseManifest.Digest}
	headBinding := semantic.SnapshotDigests{Source: rawProductionDigest(head.Digest), Semantic: headManifest.Digest}
	plannerPath, proofPath, inferencePath, evidenceIDs := productionProofPath(t, entityID, obligationID, commandID, baseBinding, headBinding)
	planInput := plannersci.Input{
		SchemaVersion: plannersci.SchemaVersion, Base: baseManifest, Head: headManifest, Registry: registry,
		CPUCapacity:     1000,
		Receipts:        []plannersci.Receipt{{CommandID: commandID, SnapshotDigest: headManifest.Digest, Envelope: productionEnvelope()}},
		ProvenancePaths: []plannersci.ProvenancePath{plannerPath},
	}
	plan := plannersci.Plan(planInput)
	if plan.Status != plannersci.StatusSelective {
		t.Fatalf("production fixture plan = %#v", plan)
	}
	proofBinding := proofsci.SnapshotBinding{Base: baseBinding, Head: headBinding}
	proofReceipt := proofsci.CommandReceipt{
		CommandID: productionID(commandID), ReceiptID: productionID("urn:gooo:shadow/receipt/test"), Status: proofsci.ReceiptVerified,
		ProviderReceiptDigest: productionDigest("provider"), PhaseReceiptDigest: productionDigest("phase"), ResourceReceiptDigest: productionDigest("resource"),
		RegistryDigest: registry.Digest, PlanDigest: plan.CanonicalDigest,
	}
	proofReceipt.Digest = proofReceipt.ExpectedDigest(proofBinding)
	proofInput := proofsci.Input{
		Schema: proofsci.SchemaVersion, Snapshots: proofBinding, RegistryDigest: registry.Digest, PlanDigest: plan.CanonicalDigest,
		ChangedRootIDs: []semantic.ID{productionID(entityID)}, SelectedCommandIDs: []semantic.ID{productionID(commandID)},
		ObligationIDs: []semantic.ID{productionID(obligationID)}, Paths: []proofsci.Path{proofPath}, CommandReceipts: []proofsci.CommandReceipt{proofReceipt},
		EvidenceIDs: evidenceIDs, InferencePath: inferencePath,
	}
	laneInput := lanesci.Input{
		SchemaVersion: lanesci.SchemaVersion, RegistryDigest: registry.Digest,
		BaseSHA: strings.Repeat("a", 40), LaneHeadSHA: strings.Repeat("b", 40), LaneID: "urn:gooo:shadow/lane/main",
		RegisteredBranch: "agent/cli-check-current2", OwnedPathPrefixes: []string{"cmd/gooo"}, ChangedPaths: []string{"cmd/gooo/selective_ci_shadow.go"},
	}
	fixture := productionFixture{base: base, head: head, planInput: planInput, proofInput: proofInput, laneInput: laneInput,
		entityID: entityID, otherID: otherID, commandID: commandID, sourceBase: sourceBase}
	fixture.reencode(t)
	return fixture
}
