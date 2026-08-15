package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

type shadowMapReader map[string][]byte

func (r shadowMapReader) ReadFile(name string) ([]byte, error) {
	data, ok := r[name]
	if !ok {
		return nil, errors.New("missing fixture file")
	}
	return append([]byte{}, data...), nil
}

type shadowFixture struct {
	files         map[string][]byte
	base          analyzersci.Snapshot
	head          analyzersci.Snapshot
	planInput     plannersci.Input
	proofInput    proofsci.Input
	laneInput     lanesci.Input
	entityID      string
	otherID       string
	commandID     string
	commandCPU    uint64
	commandMemory uint64
	sourceBase    string
}

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

func buildAnalyzerShadowSnapshot(t *testing.T, source, id, registryDigest string) analyzersci.Snapshot {
	t.Helper()
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: "cmd/gooo/fixture.gooo", PackagePath: "fixture", Source: []byte(source)}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semantic binding fixture = %#v, err = %v", result, err)
	}
	snapshot, err := analyzersci.Build(analyzersci.SnapshotInput{
		Sources:         []analyzersci.SourceInput{{Path: "cmd/gooo/fixture.gooo", BlobDigest: prefixedShadowDigest(source), Bindings: result.Bindings}},
		SourceMapDigest: prefixedShadowDigest("source-map"), RegistryDigest: registryDigest, RegisteredIDs: []string{id},
	})
	if err != nil {
		t.Fatalf("analyzer snapshot fixture = %v", err)
	}
	return snapshot
}

func shadowResourceEnvelope() resourceenvelope.Envelope {
	samples := []resourceenvelope.Sample{
		{CPUCoreNS: 1, WallNS: 10}, {CPUCoreNS: 10, WallNS: 10}, {CPUCoreNS: 20, WallNS: 10}, {CPUCoreNS: 30, WallNS: 10}, {CPUCoreNS: 40, WallNS: 10}, {CPUCoreNS: 50, WallNS: 10},
	}
	return resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: shadowDigest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5, Limits: resourceenvelope.Limits{CPUCoreNS: 100, PeakRSSBytes: 1000, ReadBytes: 1000, WriteBytes: 1000}, Samples: samples}
}

func shadowProofPath(t *testing.T, rootID, obligationID, commandID string, base, head semantic.SnapshotDigests) (plannersci.ProvenancePath, proofsci.Path, semantic.InferencePathV1, []semantic.ID) {
	t.Helper()
	root, obligation, command := commandIDToID(rootID), commandIDToID(obligationID), commandIDToID(commandID)
	receipt := commandIDToID("urn:gooo:shadow/receipt/test")
	rule := semantic.RuleBinding{ID: commandIDToID("urn:gooo:shadow/rule/v1"), Version: "1", Digest: shadowDigest("rule")}
	makeEdge := func(label string, kind semantic.InferenceKind, subject, object semantic.ID) (semantic.InferenceEdge, semantic.InferenceEvidence) {
		phase := semantic.PhasePlacement{Ordinal: 1}
		authority := semantic.AuthorityBinding{}
		controls := semantic.InferenceControls{}
		sourceBacked, independent := false, false
		switch kind {
		case semantic.InferenceAuthoritativeDeclaration:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDeclaration, semantic.AuthoritySource, semantic.AuthorityDeclare
			sourceBacked = true
		case semantic.InferenceDeterministicDerivation:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseDerivation, semantic.AuthoritySemantic, semantic.AuthorityDerive
		case semantic.InferenceIndependentVerification:
			phase.Phase, authority.Layer, authority.Effect = semantic.PhaseVerification, semantic.AuthorityVerification, semantic.AuthorityVerify
			controls.PolicyDigest, independent = shadowDigest("verification-policy"), true
		}
		evidenceID := commandIDToID("urn:gooo:shadow/evidence/" + label)
		evidenceDigest := shadowDigest("evidence/" + label)
		edge := semantic.InferenceEdge{InferenceRecord: semantic.InferenceRecord{RecordID: commandIDToID("urn:gooo:shadow/record/" + label), SubjectID: subject, ObjectID: object, Rule: rule, Phase: phase, Before: base, After: head, Authority: authority, Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}}, Controls: controls}, Kind: kind}
		if kind == semantic.InferenceAuthoritativeDeclaration {
			edge.SourceRoots = []semantic.ID{root}
		}
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: evidenceDigest, Before: base, After: head, SourceBacked: sourceBacked, Independent: independent, Controls: controls}
		return edge, evidence
	}
	first, firstEvidence := makeEdge("01-declaration", semantic.InferenceAuthoritativeDeclaration, root, obligation)
	second, secondEvidence := makeEdge("02-derivation", semantic.InferenceDeterministicDerivation, obligation, command)
	third, thirdEvidence := makeEdge("03-verification", semantic.InferenceIndependentVerification, command, receipt)
	edges := []semantic.InferenceEdge{first, second, third}
	evidence := []semantic.InferenceEvidence{firstEvidence, secondEvidence, thirdEvidence}
	recordIDs := []string{first.RecordID.String(), second.RecordID.String(), third.RecordID.String()}
	kinds := []string{string(first.Kind), string(second.Kind), string(third.Kind)}
	plannerPath := plannersci.ProvenancePath{CommandID: commandID, Path: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}, Requirement: plannersci.PathRequirement{PathID: "urn:gooo:shadow/path/main", RecordIDs: recordIDs, ExpectedKinds: kinds, StartID: rootID, EndID: receipt.String()}}
	proofPath := proofsci.Path{PathID: commandIDToID("urn:gooo:shadow/path/main"), RootID: root, ObligationID: obligation, CommandID: command, ReceiptID: receipt, RecordIDs: []semantic.ID{first.RecordID, second.RecordID, third.RecordID}, ExpectedKinds: []semantic.InferenceKind{first.Kind, second.Kind, third.Kind}}
	evidenceIDs := []semantic.ID{firstEvidence.ID, secondEvidence.ID, thirdEvidence.ID}
	return plannerPath, proofPath, semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion, Edges: edges, Evidence: evidence}, evidenceIDs
}

func commandIDToID(value string) semantic.ID { return semantic.MustIdentity(value) }

func shadowDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func prefixedShadowDigest(value string) string { return "sha256:" + shadowDigest(value) }

func (f *shadowFixture) reencodeAll() {
	f.files = map[string][]byte{}
	f.files["base_snapshot.json"], _ = f.base.CanonicalJSON()
	f.files["head_snapshot.json"], _ = f.head.CanonicalJSON()
	f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
	f.reencodeProof()
	f.reencodeLane()
}

func (f *shadowFixture) reencodeProof() {
	f.files["evidence_input.json"], _ = proofsci.EncodeInput(f.proofInput)
}

func (f *shadowFixture) reencodeLane() {
	f.files["lane_input.json"], _ = lanesci.EncodeInputJSON(f.laneInput)
}

func (f shadowFixture) reader() shadowMapReader { return shadowMapReader(f.files) }

func (f shadowFixture) args() []string {
	return []string{"shadow", "--base-snapshot", "base_snapshot.json", "--head-snapshot", "head_snapshot.json", "--plan-input", "plan_input.json", "--evidence-input", "evidence_input.json", "--lane-input", "lane_input.json"}
}

func (f shadowFixture) argsReversed() []string {
	return []string{"shadow", "--lane-input", "lane_input.json", "--evidence-input", "evidence_input.json", "--plan-input", "plan_input.json", "--head-snapshot", "head_snapshot.json", "--base-snapshot", "base_snapshot.json"}
}

func (f *shadowFixture) reverseInputs() {
	f.planInput.Registry.Commands = append([]plannersci.Command(nil), f.planInput.Registry.Commands...)
	if len(f.proofInput.InferencePath.Edges) > 1 {
		for left, right := 0, len(f.proofInput.InferencePath.Edges)-1; left < right; left, right = left+1, right-1 {
			f.proofInput.InferencePath.Edges[left], f.proofInput.InferencePath.Edges[right] = f.proofInput.InferencePath.Edges[right], f.proofInput.InferencePath.Edges[left]
		}
	}
	if len(f.proofInput.InferencePath.Evidence) > 1 {
		for left, right := 0, len(f.proofInput.InferencePath.Evidence)-1; left < right; left, right = left+1, right-1 {
			f.proofInput.InferencePath.Evidence[left], f.proofInput.InferencePath.Evidence[right] = f.proofInput.InferencePath.Evidence[right], f.proofInput.InferencePath.Evidence[left]
		}
	}
	f.laneInput.OwnedPathPrefixes = append([]string(nil), f.laneInput.OwnedPathPrefixes...)
	f.laneInput.ChangedPaths = append([]string(nil), f.laneInput.ChangedPaths...)
	f.reencodeAll()
}
