package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	lanesci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/lanefrontier"
	proofsci "github.com/kimjooyoon/meta-ontology-go/internal/provenance/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

type productionFixture struct {
	files      map[string][]byte
	base       analyzersci.Snapshot
	head       analyzersci.Snapshot
	planInput  plannersci.Input
	proofInput proofsci.Input
	laneInput  lanesci.Input
	entityID   string
	otherID    string
	commandID  string
	sourceBase string
}

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

func buildProductionSnapshot(t *testing.T, source, id, registryDigest string) analyzersci.Snapshot {
	t.Helper()
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: "cmd/gooo/fixture.gooo", PackagePath: "fixture", Source: []byte(source)}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semantic binding fixture = %#v, err = %v", result, err)
	}
	snapshot, err := analyzersci.Build(analyzersci.SnapshotInput{
		Sources:         []analyzersci.SourceInput{{Path: "cmd/gooo/fixture.gooo", BlobDigest: productionPrefixedDigest(source), Bindings: result.Bindings}},
		SourceMapDigest: productionPrefixedDigest("source-map"), RegistryDigest: registryDigest, RegisteredIDs: []string{id},
	})
	if err != nil {
		t.Fatalf("analyzer snapshot fixture = %v", err)
	}
	return snapshot
}

func productionManifest(t *testing.T, snapshot analyzersci.Snapshot) plannersci.SnapshotManifest {
	t.Helper()
	files := make([]plannersci.SnapshotFile, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		ids := make([]string, 0, len(source.Bindings))
		for _, binding := range source.Bindings {
			ids = append(ids, binding.ID)
		}
		files = append(files, plannersci.SnapshotFile{Path: source.Path, BlobDigest: rawProductionDigest(source.BlobDigest), SemanticIDs: ids})
	}
	manifest := plannersci.SnapshotManifest{SchemaVersion: plannersci.ManifestSchemaVersion, Files: files}
	manifest.Digest = manifest.ComputedDigest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("planner manifest fixture = %v", err)
	}
	return manifest
}

func productionEnvelope() resourceenvelope.Envelope {
	samples := []resourceenvelope.Sample{{CPUCoreNS: 1, WallNS: 10}, {CPUCoreNS: 10, WallNS: 10}, {CPUCoreNS: 20, WallNS: 10}, {CPUCoreNS: 30, WallNS: 10}, {CPUCoreNS: 40, WallNS: 10}, {CPUCoreNS: 50, WallNS: 10}}
	return resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: productionDigest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5,
		Limits: resourceenvelope.Limits{CPUCoreNS: 100, PeakRSSBytes: 1000, ReadBytes: 1000, WriteBytes: 1000}, Samples: samples}
}

func productionID(value string) semantic.ID { return semantic.MustIdentity(value) }

func productionDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func productionPrefixedDigest(value string) string { return "sha256:" + productionDigest(value) }

func rawProductionDigest(value string) string { return strings.TrimPrefix(value, "sha256:") }
