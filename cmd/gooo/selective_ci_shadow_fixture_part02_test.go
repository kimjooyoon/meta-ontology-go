package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func newShadowFixture(t *testing.T) shadowFixture {
	t.Helper()
	entityID := "urn:gooo:shadow/entity/order"
	otherID := "urn:gooo:shadow/entity/other"
	commandID := "urn:gooo:shadow/command/test"
	obligationID := "urn:gooo:shadow/obligation/test"
	sourceBase := `package fixture

//gooo:bind id="urn:gooo:shadow/entity/order" role="HANDWRITTEN_IMPL"
func Order() {}
`
	commandCPU, commandMemory := uint64(100), uint64(1000)
	registry := plannersci.Registry{
		SchemaVersion:       plannersci.RegistrySchemaVersion,
		PolicyDigest:        shadowDigest("policy"),
		Nodes:               []impactgraph.Node{{ID: entityID, Kind: impactgraph.NodeKindSemantic}},
		DependencyEdges:     []plannersci.DependencyEdge{},
		Obligations:         []plannersci.ObligationBinding{{ID: obligationID, Subject: entityID, CommandIDs: []string{commandID}}},
		Commands:            []plannersci.Command{{ID: commandID, Argv: []string{"gooo-shadow-sentinel", "never-run"}, WorkingDir: ".", CPUWorkUnits: commandCPU, MemoryBytes: commandMemory}},
		GlobalGuardCommands: []plannersci.Command{},
	}
	registry.Digest = registry.ComputedDigest()
	base := buildAnalyzerShadowSnapshot(t, sourceBase+"// base\n", entityID, "sha256:"+registry.Digest)
	head := buildAnalyzerShadowSnapshot(t, sourceBase+"// head\n", entityID, "sha256:"+registry.Digest)
	baseManifest, err := plannerManifestFromAnalyzerSnapshot(base)
	if err != nil {
		t.Fatal(err)
	}
	headManifest, err := plannerManifestFromAnalyzerSnapshot(head)
	if err != nil {
		t.Fatal(err)
	}
	baseSnapshot := semantic.SnapshotDigests{Source: rawDigest(base.Digest), Semantic: baseManifest.Digest}
	headSnapshot := semantic.SnapshotDigests{Source: rawDigest(head.Digest), Semantic: headManifest.Digest}
	plannerPath, proofPath, inferencePath, evidenceIDs := shadowProofPath(t, entityID, obligationID, commandID, baseSnapshot, headSnapshot)
	plannerInput := plannersci.Input{
		SchemaVersion:   plannersci.SchemaVersion,
		Base:            baseManifest,
		Head:            headManifest,
		Registry:        registry,
		CPUCapacity:     1000,
		Receipts:        []plannersci.Receipt{{CommandID: commandID, SnapshotDigest: headManifest.Digest, Envelope: shadowResourceEnvelope()}},
		ProvenancePaths: []plannersci.ProvenancePath{plannerPath},
	}
	plan := plannersci.Plan(plannerInput)
	if plan.Status != plannersci.StatusSelective {
		t.Fatalf("fixture plan = %#v", plan)
	}
	proofBinding := proofsci.SnapshotBinding{Base: baseSnapshot, Head: headSnapshot}
	proofReceipt := proofsci.CommandReceipt{
		CommandID: commandIDToID(commandID), ReceiptID: commandIDToID("urn:gooo:shadow/receipt/test"), Status: proofsci.ReceiptVerified,
		ProviderReceiptDigest: shadowDigest("provider"), PhaseReceiptDigest: shadowDigest("phase"), ResourceReceiptDigest: shadowDigest("resource"),
		RegistryDigest: registry.Digest, PlanDigest: plan.CanonicalDigest,
	}
	proofReceipt.Digest = proofReceipt.ExpectedDigest(proofBinding)
	proofInput := proofsci.Input{
		Schema: proofsci.SchemaVersion, Snapshots: proofBinding, RegistryDigest: registry.Digest, PlanDigest: plan.CanonicalDigest,
		ChangedRootIDs: []semantic.ID{commandIDToID(entityID)}, SelectedCommandIDs: []semantic.ID{commandIDToID(commandID)},
		ObligationIDs: []semantic.ID{commandIDToID(obligationID)}, Paths: []proofsci.Path{proofPath}, CommandReceipts: []proofsci.CommandReceipt{proofReceipt},
		EvidenceIDs: evidenceIDs, InferencePath: inferencePath,
	}
	laneInput := lanesci.Input{
		SchemaVersion: lanesci.SchemaVersion, RegistryDigest: registry.Digest, BaseSHA: strings.Repeat("a", 40), LaneHeadSHA: strings.Repeat("b", 40),
		LaneID: "urn:gooo:shadow/lane/main", RegisteredBranch: "agent/cli-check-current2", OwnedPathPrefixes: []string{"cmd/gooo"}, ChangedPaths: []string{"cmd/gooo/selective_ci_shadow.go"},
	}
	fixture := shadowFixture{base: base, head: head, planInput: plannerInput, proofInput: proofInput, laneInput: laneInput, entityID: entityID, otherID: otherID, commandID: commandID, commandCPU: commandCPU, commandMemory: commandMemory, sourceBase: sourceBase}
	fixture.reencodeAll()
	return fixture
}
