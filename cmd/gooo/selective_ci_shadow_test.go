package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
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

func TestSelectiveCIShadowPositiveSelfDigestAndNoExecution(t *testing.T) {
	fixture := newShadowFixture(t)
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("shadow stderr = %q", stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.Status != "SHADOW_SELECTIVE" || output.Stage != "SELECTIVE" || output.Reason != "VERIFIED" {
		t.Fatalf("shadow output classification = %#v", output)
	}
	if output.ExecutionAuthorized || !output.ShadowOnly {
		t.Fatalf("shadow execution flags = authorized:%t shadow_only:%t", output.ExecutionAuthorized, output.ShadowOnly)
	}
	if len(output.SelectedCommands) != 1 || output.SelectedCommands[0].ID != fixture.commandID || !reflect.DeepEqual(output.SelectedCommands[0].Argv, []string{"gooo-shadow-sentinel", "never-run"}) {
		t.Fatalf("selected command projection = %#v", output.SelectedCommands)
	}
	if len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 1 || len(output.ResourceReceipts) != 1 {
		t.Fatalf("selected receipt projection = guards:%#v work:%#v receipts:%#v", output.SelectedGuards, output.SelectedWorkIDs, output.ResourceReceipts)
	}
	if output.ResourceReceipts[0].CPUWorkUnits != fixture.commandCPU || output.ResourceReceipts[0].MemoryBytes != fixture.commandMemory {
		t.Fatalf("resource receipt = %#v", output.ResourceReceipts[0])
	}
	if output.Lane.BaseSHA != fixture.laneInput.BaseSHA || output.Lane.LaneHeadSHA != fixture.laneInput.LaneHeadSHA || output.Lane.LaneID != fixture.laneInput.LaneID || output.Lane.Reason != string(lanesci.ReasonEligible) {
		t.Fatalf("lane projection = %#v", output.Lane)
	}
	if output.CanonicalDigest == "" || output.CanonicalDigest != output.stableDigest() {
		t.Fatalf("output self digest = %q", output.CanonicalDigest)
	}
	t.Logf("canonical receipt digest=%s", output.CanonicalDigest)
	if bytes.Contains(stdout.Bytes(), []byte(`"execution_authorized":true`)) {
		t.Fatal("shadow receipt authorized execution")
	}
}

func TestSelectiveCIShadowInputPermutationIsByteStable(t *testing.T) {
	left := newShadowFixture(t)
	right := newShadowFixture(t)
	right.reverseInputs()
	var leftOut, leftErr, rightOut, rightErr bytes.Buffer
	if code := runSelectiveCI(left.args(), left.reader(), &leftOut, &leftErr); code != exitOK {
		t.Fatalf("left code = %d, stderr = %q", code, leftErr.String())
	}
	if code := runSelectiveCI(right.argsReversed(), right.reader(), &rightOut, &rightErr); code != exitOK {
		t.Fatalf("right code = %d, stderr = %q", code, rightErr.String())
	}
	if !bytes.Equal(leftOut.Bytes(), rightOut.Bytes()) {
		t.Fatalf("input permutation changed receipt:\nleft=%s\nright=%s", leftOut.Bytes(), rightOut.Bytes())
	}
}

func TestSelectiveCIShadowUsageRejectsUnknownDuplicateMissingAndTrailing(t *testing.T) {
	fixture := newShadowFixture(t)
	cases := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: append(fixture.args(), "--unknown", "value")},
		{name: "duplicate flag", args: append(fixture.args(), "--lane-input", "lane.json")},
		{name: "missing value", args: []string{"shadow", "--base-snapshot"}},
		{name: "trailing positional", args: append(fixture.args(), "trailing")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := runSelectiveCI(test.args, fixture.reader(), &stdout, &stderr); code != exitUsage || stdout.Len() != 0 || !strings.Contains(stderr.String(), "usage: gooo selective-ci shadow") {
				t.Fatalf("usage result = code:%d stdout:%q stderr:%q", code, stdout.String(), stderr.String())
			}
		})
	}

	for _, name := range []string{"base_snapshot", "head_snapshot", "plan_input", "evidence_input", "lane_input"} {
		t.Run("missing "+name, func(t *testing.T) {
			missing := newShadowFixture(t)
			delete(missing.files, name+".json")
			var stdout, stderr bytes.Buffer
			if code := runSelectiveCI(missing.args(), missing.reader(), &stdout, &stderr); code != exitUsage || stdout.Len() != 0 || !strings.Contains(stderr.String(), "cli.usage") {
				t.Fatalf("missing result = code:%d stdout:%q stderr:%q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestSelectiveCIShadowMalformedUnknownAndDuplicateJSONFallback(t *testing.T) {
	cases := []struct {
		name      string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "malformed", component: "evidence_input", mutate: func(f *shadowFixture) { f.files["evidence_input.json"] = []byte("{") }},
		{name: "unknown field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = append(f.files["evidence_input.json"], []byte(" ")...)
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("{"), []byte("{\"unknown\":true,"), 1)
		}},
		{name: "duplicate field", component: "evidence_input", mutate: func(f *shadowFixture) {
			f.files["evidence_input.json"] = bytes.Replace(f.files["evidence_input.json"], []byte("\"schema\":"), []byte("\"schema\":\"gooo-selective-ci-evidence/v1\",\"schema\":"), 1)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != "INPUT" || output.Component != test.component || output.Reason == "" {
				t.Fatalf("fallback = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowBindingAndPlannerPrecedence(t *testing.T) {
	cases := []struct {
		name      string
		stage     string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "stale snapshot", stage: "INPUT", component: "base_snapshot", mutate: func(f *shadowFixture) {
			f.files["base_snapshot.json"] = bytes.Replace(f.files["base_snapshot.json"], []byte("sha256:"), []byte("sha256:0"), 1)
		}},
		{name: "manifest mismatch", stage: "SNAPSHOT_BINDING", component: "base_manifest", mutate: func(f *shadowFixture) {
			f.planInput.Base.Files[0].SemanticIDs = []string{f.entityID, f.otherID}
			f.planInput.Base.Digest = f.planInput.Base.ComputedDigest()
			f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
		}},
		{name: "registry mismatch", stage: "REGISTRY_BINDING", component: "base_snapshot", mutate: func(f *shadowFixture) {
			f.base = buildAnalyzerShadowSnapshot(t, f.sourceBase+"// base\n", f.entityID, prefixedShadowDigest("different-registry"))
			f.files["base_snapshot.json"], _ = f.base.CanonicalJSON()
		}},
		{name: "planner fallback", stage: "PLAN", component: "planner", mutate: func(f *shadowFixture) {
			f.planInput.CPUCapacity = 1
			f.files["plan_input.json"], _ = plannersci.EncodeJSON(f.planInput)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != test.stage || output.Component != test.component || output.ExecutionAuthorized || !output.ShadowOnly {
				t.Fatalf("fallback = %#v", output)
			}
			if len(output.SelectedCommands) != 0 || len(output.SelectedGuards) != 0 || len(output.SelectedWorkIDs) != 0 {
				t.Fatalf("fallback exposed selection = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowProofBindingFailurePrecedence(t *testing.T) {
	cases := []struct {
		name      string
		component string
		mutate    func(*shadowFixture)
	}{
		{name: "registry", component: "registry_digest", mutate: func(f *shadowFixture) {
			f.proofInput.RegistryDigest = shadowDigest("wrong-proof-registry")
			f.reencodeProof()
		}},
		{name: "plan", component: "plan_digest", mutate: func(f *shadowFixture) { f.proofInput.PlanDigest = shadowDigest("wrong-proof-plan"); f.reencodeProof() }},
		{name: "changed roots", component: "changed_root_ids", mutate: func(f *shadowFixture) {
			f.proofInput.ChangedRootIDs = []semantic.ID{commandIDToID(f.otherID)}
			f.reencodeProof()
		}},
		{name: "selected commands", component: "selected_command_ids", mutate: func(f *shadowFixture) {
			f.proofInput.SelectedCommandIDs = []semantic.ID{commandIDToID(f.otherID)}
			f.reencodeProof()
		}},
		{name: "snapshots", component: "snapshots", mutate: func(f *shadowFixture) {
			f.proofInput.Snapshots.Head.Semantic = shadowDigest("wrong-proof-snapshot")
			f.reencodeProof()
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Stage != "PLAN_PROOF_BINDING" || output.Component != test.component || output.Status != "FULL_SUITE_FALLBACK" {
				t.Fatalf("proof binding fallback = %#v", output)
			}
		})
	}
}

func TestSelectiveCIShadowProofAndLaneFailClosedPartitions(t *testing.T) {
	cases := []struct {
		name   string
		stage  string
		mutate func(*shadowFixture)
	}{
		{name: "proof fail closed", stage: "PROOF_FAIL_CLOSED", mutate: func(f *shadowFixture) {
			f.proofInput.CommandReceipts[0].Digest = shadowDigest("wrong-receipt")
			f.reencodeProof()
		}},
		{name: "proof unknown", stage: "PROOF_UNKNOWN", mutate: func(f *shadowFixture) {
			f.proofInput.InferencePath.Edges = f.proofInput.InferencePath.Edges[:2]
			f.reencodeProof()
		}},
		{name: "lane unknown", stage: "LANE_UNKNOWN", mutate: func(f *shadowFixture) { f.laneInput.BaseSHA = ""; f.reencodeLane() }},
		{name: "lane ineligible", stage: "LANE_INELIGIBLE", mutate: func(f *shadowFixture) { f.laneInput.ActiveLeaseCount = 1; f.reencodeLane() }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := newShadowFixture(t)
			test.mutate(&fixture)
			output := runShadowFixture(t, fixture)
			if output.Status != "FULL_SUITE_FALLBACK" || output.Stage != test.stage || output.ExecutionAuthorized || !output.ShadowOnly {
				t.Fatalf("partition fallback = %#v", output)
			}
		})
	}
}

func runShadowFixture(t *testing.T, fixture shadowFixture) selectiveCIShadowOutput {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if code := runSelectiveCI(fixture.args(), fixture.reader(), &stdout, &stderr); code != exitOK {
		t.Fatalf("shadow code = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	var output selectiveCIShadowOutput
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &output); err != nil {
		t.Fatalf("decode shadow output: %v", err)
	}
	if output.CanonicalDigest == "" || output.CanonicalDigest != output.stableDigest() {
		t.Fatalf("invalid output digest: %#v", output)
	}
	return output
}

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
		SchemaVersion: plannersci.SchemaVersion,
		Base:          baseManifest, Head: headManifest, Registry: registry, CPUCapacity: 1000,
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
