package selectiveci

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func testID(value string) semantic.ID { return semantic.MustIdentity("selective-ci-test://" + value) }

func testDigest(value string) string { return semantic.StableHashString("selective-ci-test/" + value) }

type testFixture struct {
	input      Input
	root       semantic.ID
	obligation semantic.ID
	command    semantic.ID
	receipt    semantic.ID
	branch     semantic.InferenceEdge
}

func completeFixture() testFixture {
	base := semantic.SnapshotDigests{Source: testDigest("base-source"), Semantic: testDigest("base-semantic")}
	head := semantic.SnapshotDigests{Source: testDigest("head-source"), Semantic: testDigest("head-semantic")}
	binding := SnapshotBinding{Base: base, Head: head}
	root, obligation := testID("root"), testID("obligation")
	command, receipt := testID("command"), testID("receipt")
	registry, plan := testDigest("registry"), testDigest("plan")
	rule := semantic.RuleBinding{ID: testID("rule"), Version: "1", Digest: testDigest("rule")}
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
			controls.PolicyDigest, independent = testDigest("verification-policy"), true
		default:
			panic("unsupported fixture kind")
		}
		evidenceID, evidenceDigest := testID("evidence/"+label), testDigest("evidence/"+label)
		edge := semantic.InferenceEdge{InferenceRecord: semantic.InferenceRecord{
			RecordID: testID("record/" + label), SubjectID: subject, ObjectID: object, Rule: rule,
			Phase: phase, Before: base, After: head, Authority: authority,
			Evidence: []semantic.EvidenceReference{{ID: evidenceID, Digest: evidenceDigest}}, Controls: controls,
		}, Kind: kind}
		if kind == semantic.InferenceAuthoritativeDeclaration {
			edge.SourceRoots = []semantic.ID{root}
		}
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: evidenceDigest, Before: base, After: head, SourceBacked: sourceBacked, Independent: independent, Controls: controls}
		return edge, evidence
	}
	first, firstEvidence := makeEdge("declaration", semantic.InferenceAuthoritativeDeclaration, root, obligation)
	second, secondEvidence := makeEdge("derivation", semantic.InferenceDeterministicDerivation, obligation, command)
	third, thirdEvidence := makeEdge("verification", semantic.InferenceIndependentVerification, command, receipt)
	path := Path{PathID: testID("path/main"), RootID: root, ObligationID: obligation, CommandID: command, ReceiptID: receipt,
		RecordIDs: []semantic.ID{first.RecordID, second.RecordID, third.RecordID}, ExpectedKinds: []semantic.InferenceKind{first.Kind, second.Kind, third.Kind}}
	commandReceipt := CommandReceipt{CommandID: command, ReceiptID: receipt, Status: ReceiptVerified,
		ProviderReceiptDigest: testDigest("provider"), PhaseReceiptDigest: testDigest("phase"), ResourceReceiptDigest: testDigest("resource"), RegistryDigest: registry, PlanDigest: plan}
	commandReceipt.Digest = commandReceipt.ExpectedDigest(binding)
	return testFixture{input: Input{Schema: SchemaVersion, Snapshots: binding, RegistryDigest: registry, PlanDigest: plan,
		ChangedRootIDs: []semantic.ID{root}, SelectedCommandIDs: []semantic.ID{command}, ObligationIDs: []semantic.ID{obligation}, Paths: []Path{path}, CommandReceipts: []CommandReceipt{commandReceipt},
		EvidenceIDs: []semantic.ID{firstEvidence.ID, secondEvidence.ID, thirdEvidence.ID}, InferencePath: semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion,
			Edges: []semantic.InferenceEdge{first, second, third}, Evidence: []semantic.InferenceEvidence{firstEvidence, secondEvidence, thirdEvidence}}}, root: root, obligation: obligation, command: command, receipt: receipt}
}

func cloneFixture(fixture testFixture) testFixture {
	copy := fixture
	copy.input.ChangedRootIDs = append([]semantic.ID(nil), fixture.input.ChangedRootIDs...)
	copy.input.SelectedCommandIDs = append([]semantic.ID(nil), fixture.input.SelectedCommandIDs...)
	copy.input.ObligationIDs = append([]semantic.ID(nil), fixture.input.ObligationIDs...)
	copy.input.Paths = append([]Path(nil), fixture.input.Paths...)
	copy.input.CommandReceipts = append([]CommandReceipt(nil), fixture.input.CommandReceipts...)
	copy.input.EvidenceIDs = append([]semantic.ID(nil), fixture.input.EvidenceIDs...)
	copy.input.InferencePath.Edges = append([]semantic.InferenceEdge(nil), fixture.input.InferencePath.Edges...)
	copy.input.InferencePath.Claims = append([]semantic.SemanticChangeClaim(nil), fixture.input.InferencePath.Claims...)
	copy.input.InferencePath.Evidence = append([]semantic.InferenceEvidence(nil), fixture.input.InferencePath.Evidence...)
	for index := range copy.input.InferencePath.Edges {
		copy.input.InferencePath.Edges[index].Evidence = append([]semantic.EvidenceReference(nil), fixture.input.InferencePath.Edges[index].Evidence...)
		copy.input.InferencePath.Edges[index].SourceRoots = append([]semantic.ID(nil), fixture.input.InferencePath.Edges[index].SourceRoots...)
	}
	return copy
}

func TestEvaluateSelectiveCIClosurePartitions(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*testFixture)
		status DecisionStatus
		code   string
	}{
		{name: "complete", status: Verified, code: CodeVerified},
		{name: "missing record", mutate: func(f *testFixture) { f.input.InferencePath.Edges = f.input.InferencePath.Edges[:2] }, status: Unknown, code: CodeMissing},
		{name: "duplicate path", mutate: func(f *testFixture) { f.input.Paths = append(f.input.Paths, f.input.Paths[0]) }, status: FailClosed, code: CodeDuplicate},
		{name: "ambiguity", mutate: func(f *testFixture) {
			branch := f.input.InferencePath.Edges[1]
			branch.RecordID, branch.ObjectID = testID("record/branch"), testID("branch-command")
			f.branch = branch
			f.input.InferencePath.Edges = append(f.input.InferencePath.Edges, branch)
			f.input.Paths[0].RecordIDs = append(f.input.Paths[0].RecordIDs, branch.RecordID)
			f.input.Paths[0].ExpectedKinds = append(f.input.Paths[0].ExpectedKinds, branch.Kind)
		}, status: FailClosed, code: CodeAmbiguous},
		{name: "cycle", mutate: func(f *testFixture) { f.input.InferencePath.Edges[2].ObjectID = f.obligation }, status: FailClosed, code: CodeCycle},
		{name: "wrong endpoint", mutate: func(f *testFixture) { f.input.Paths[0].ReceiptID = testID("receipt/wrong") }, status: FailClosed, code: CodeWrongEndpoint},
		{name: "stale snapshot", mutate: func(f *testFixture) { f.input.Snapshots.Head.Semantic = testDigest("stale-head") }, status: Unknown, code: CodeStaleSnapshot},
		{name: "candidate only", mutate: func(f *testFixture) {
			edge := &f.input.InferencePath.Edges[1]
			edge.Kind = semantic.InferenceObservationCandidate
			edge.Phase.Phase = semantic.PhaseObservation
			edge.Authority = semantic.AuthorityBinding{Layer: semantic.AuthorityCandidate, Effect: semantic.AuthorityObserve}
			edge.Controls.CatalogDigest = testDigest("catalog")
			edge.Before.Semantic = edge.After.Semantic
			f.input.InferencePath.Evidence[1].Before = edge.Before
			f.input.InferencePath.Evidence[1].After = edge.After
			f.input.InferencePath.Evidence[1].Controls = edge.Controls
		}, status: Unknown, code: CodeCandidate},
		{name: "receipt mismatch", mutate: func(f *testFixture) { f.input.CommandReceipts[0].Digest = testDigest("wrong-receipt") }, status: FailClosed, code: CodeReceiptMismatch},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fixture := cloneFixture(completeFixture())
			if test.mutate != nil {
				test.mutate(&fixture)
			}
			got := Evaluate(fixture.input)
			if got.Status != test.status || got.Code != test.code {
				t.Fatalf("receipt status/code = %s/%s, want %s/%s: %#v", got.Status, got.Code, test.status, test.code, got)
			}
			if got.Status == Verified && got.Fallback != NoFallback {
				t.Fatalf("verified fallback mode = %q", got.Fallback)
			}
			if got.Status != Verified && (got.Fallback != FullSuite || got.VerifiedCommandCount != 0) {
				t.Fatalf("fallback receipt = %#v", got)
			}
		})
	}
}

func TestEvaluateReceiptCountsAndDigest(t *testing.T) {
	got := Evaluate(completeFixture().input)
	if got.Status != Verified || got.Fallback != NoFallback || got.VerifiedCommandCount != 1 || got.VerifiedObligationCount != 1 || got.VerifiedPathCount != 1 {
		t.Fatalf("complete receipt = %#v", got)
	}
	if got.Digest != got.ExpectedDigest() || got.Canonical() == "" {
		t.Fatalf("receipt digest/canonical mismatch: %#v", got)
	}
}

func TestPermutationsHaveIdenticalCanonicalReceipt(t *testing.T) {
	left := Evaluate(completeFixture().input)
	rightFixture := cloneFixture(completeFixture())
	reverseEdges := append([]semantic.InferenceEdge(nil), rightFixture.input.InferencePath.Edges...)
	for leftIndex, rightIndex := 0, len(reverseEdges)-1; leftIndex < rightIndex; leftIndex, rightIndex = leftIndex+1, rightIndex-1 {
		reverseEdges[leftIndex], reverseEdges[rightIndex] = reverseEdges[rightIndex], reverseEdges[leftIndex]
	}
	rightFixture.input.InferencePath.Edges = reverseEdges
	reverseEvidence := append([]semantic.InferenceEvidence(nil), rightFixture.input.InferencePath.Evidence...)
	for leftIndex, rightIndex := 0, len(reverseEvidence)-1; leftIndex < rightIndex; leftIndex, rightIndex = leftIndex+1, rightIndex-1 {
		reverseEvidence[leftIndex], reverseEvidence[rightIndex] = reverseEvidence[rightIndex], reverseEvidence[leftIndex]
	}
	rightFixture.input.InferencePath.Evidence = reverseEvidence
	right := Evaluate(rightFixture.input)
	if left.Canonical() != right.Canonical() || left.Digest != right.Digest {
		t.Fatalf("permutation changed receipt:\nleft=%s\nright=%s", left.Canonical(), right.Canonical())
	}
}

func TestPermutationProperty(t *testing.T) {
	property := func(keys []uint8) bool {
		fixture := cloneFixture(completeFixture())
		order := []uint8{0, 0, 0}
		copy(order, keys)
		sort.SliceStable(fixture.input.InferencePath.Edges, func(left, right int) bool {
			return order[left] < order[right]
		})
		sort.SliceStable(fixture.input.InferencePath.Evidence, func(left, right int) bool {
			return order[left] < order[right]
		})
		return Evaluate(fixture.input).Canonical() == Evaluate(completeFixture().input).Canonical()
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 64}); err != nil {
		t.Fatal(err)
	}
}

func TestStrictJSONCodecsRoundTripAndRejectExtraFields(t *testing.T) {
	input := completeFixture().input
	data, err := EncodeInput(input)
	if err != nil {
		t.Fatalf("encode input: %v", err)
	}
	decoded, err := DecodeInput(data)
	if err != nil {
		t.Fatalf("decode input: %v", err)
	}
	if got := Evaluate(decoded); got.Status != Verified {
		t.Fatalf("decoded input receipt = %#v", got)
	}
	receipt := Evaluate(input)
	receiptData, err := EncodeReceipt(receipt)
	if err != nil {
		t.Fatalf("encode receipt: %v", err)
	}
	decodedReceipt, err := DecodeReceipt(receiptData)
	if err != nil || decodedReceipt.Digest != receipt.Digest {
		t.Fatalf("receipt round trip = %#v, err=%v", decodedReceipt, err)
	}
	var wire map[string]any
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	wire["unexpected"] = true
	withExtra, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeInput(withExtra); err == nil {
		t.Fatal("unknown input field was accepted")
	}
	duplicate := bytes.Replace(data, []byte(`"schema"`), []byte(`"schema":"`+SchemaVersion+`","schema"`), 1)
	if _, err := DecodeInput(duplicate); err == nil {
		t.Fatal("duplicate input field was accepted")
	}
}

func TestReceiptJSONHasCanonicalOrdering(t *testing.T) {
	receipt := Evaluate(completeFixture().input)
	first, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(receipt)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("receipt JSON is not deterministic: %s / %s", first, second)
	}
}
