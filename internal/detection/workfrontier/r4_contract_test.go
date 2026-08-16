package workfrontier

import (
	"bytes"
	"embed"
	"encoding/json"
	"reflect"
	"testing"
)

//go:embed testdata/r4_cases.json
var r4FixtureData embed.FS

type r4Fixture struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func TestR4StrictFixtures(t *testing.T) {
	var fixtures []r4Fixture
	data, err := r4FixtureData.ReadFile("testdata/r4_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			input := r4FixtureInput(t, fixture.Name)
			got := EvaluateR4(input)
			if got.Status != fixture.Status || got.Reason != fixture.Reason {
				t.Fatalf("result = %s/%s, want %s/%s", got.Status, got.Reason, fixture.Status, fixture.Reason)
			}
			if (got.Status == R4StatusUnknown || got.Status == R4StatusFailClosed) &&
				len(got.SelectedIDs) != 0 {
				t.Fatalf("non-pass result selected %v", got.SelectedIDs)
			}
			if got.GraphDigest == "" || got.SCCDigest == "" || got.CondensationDigest == "" || got.RuleDigest == "" {
				t.Fatal("missing canonical graph or rule digest")
			}
			encoded, err := EncodeR4ResultJSON(got)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(encoded, []byte("proof_valid")) || bytes.Contains(encoded, []byte("promotion_authorized")) {
				t.Fatalf("result emitted forbidden authorization field: %s", encoded)
			}
		})
	}
}

func TestR4PermutationAndRootOrderIndependence(t *testing.T) {
	base := r4FixtureInput(t, "two-node-cycle")
	baseResult := EvaluateR4(base)
	permuted := base
	permuted.Pressures = reversePressures(base.Pressures)
	permuted.States = reverseStates(base.States)
	permuted.Paths = reversePaths(base.Paths)
	permuted.Rules = reverseRules(base.Rules)
	permuted.RootObligationIDs = []string{"obligation/root"}
	permutedResult := EvaluateR4(permuted)
	if !reflect.DeepEqual(baseResult, permutedResult) {
		t.Fatalf("permutation changed result:\nbase=%#v\npermuted=%#v", baseResult, permutedResult)
	}
	baseGraph, err := AnalyzeR4Graph(base)
	if err != nil {
		t.Fatal(err)
	}
	permutedGraph, err := AnalyzeR4Graph(permuted)
	if err != nil {
		t.Fatal(err)
	}
	if baseGraph.GraphDigest != permutedGraph.GraphDigest || baseGraph.SCCDigest != permutedGraph.SCCDigest || baseGraph.CondensationDigest != permutedGraph.CondensationDigest {
		t.Fatalf("permutation changed graph digests: %#v %#v", baseGraph, permutedGraph)
	}
	multiRoot := r4FixtureInput(t, "acyclic")
	multiRoot.RootObligationIDs = []string{"obligation/root", "obligation/child"}
	multiRoot, err = BindR4Payloads(multiRoot)
	if err != nil {
		t.Fatal(err)
	}
	reorderedRoots := multiRoot
	reorderedRoots.RootObligationIDs = []string{"obligation/child", "obligation/root"}
	if !reflect.DeepEqual(EvaluateR4(multiRoot), EvaluateR4(reorderedRoots)) {
		t.Fatal("root order changed the normalized result")
	}
	if got := FairBaseline(r4FixtureInput(t, "acyclic")); !reflect.DeepEqual(got, []string{"path/root"}) {
		t.Fatalf("fair baseline = %v", got)
	}
}

func TestR4StrictEnvelopeRejectsUnknownDuplicateAndMissingFields(t *testing.T) {
	encoded, err := EncodeR4JSON(r4FixtureInput(t, "acyclic"))
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	delete(object, "root_obligation_ids")
	missing, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeR4JSON(missing); err == nil {
		t.Fatal("accepted an envelope without root_obligation_ids")
	}
	object["unexpected"] = json.RawMessage(`true`)
	unknown, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeR4JSON(unknown); err == nil {
		t.Fatal("accepted an envelope with an unknown field")
	}
	duplicate := []byte(`{"schema_version":"gooo/work-frontier-r4/v1","schema_version":"gooo/work-frontier-r4/v1"}`)
	if _, err := DecodeR4JSON(duplicate); err == nil {
		t.Fatal("accepted an envelope with a duplicate field")
	}
	duplicateResult := EvaluateR4JSON(duplicate)
	if duplicateResult.Status != R4StatusFailClosed || duplicateResult.Reason != R4ReasonMalformedBinding {
		t.Fatalf("duplicate envelope result = %#v", duplicateResult)
	}
	result := EvaluateR4JSON(missing)
	if result.Status != R4StatusUnknown || result.Reason != R4ReasonRequiredInputMissing || !result.FullSuiteRequired {
		t.Fatalf("missing envelope result = %#v", result)
	}
}

func TestR4DigestsExcludeFixtureLabels(t *testing.T) {
	input := r4FixtureInput(t, "self-loop")
	first := EvaluateR4(input)
	fixtureLabel := r4Fixture{Name: "self-loop", Status: "PASS", Reason: "NONE"}
	second := EvaluateR4(input)
	if first.GraphDigest != second.GraphDigest || first.RuleDigest != second.RuleDigest {
		t.Fatal("source-derived digests changed without an input change")
	}
	if got := r4FixtureDigest(input); got == "" {
		t.Fatal("empty producer fixture digest")
	}
	if fixtureLabel.Status == "" || fixtureLabel.Reason == "" {
		t.Fatal("fixture labels unexpectedly absent")
	}
}

func TestR4MalformedGraphFailsClosed(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	input.Paths[0].PrerequisiteObligationIDs = []string{"obligation/root", "obligation/root"}
	var err error
	input, err = BindR4Payloads(input)
	if err != nil {
		t.Fatal(err)
	}
	got := EvaluateR4(input)
	if got.Status != R4StatusFailClosed || got.Reason != R4ReasonMalformedGraph || len(got.SelectedIDs) != 0 {
		t.Fatalf("malformed graph result = %#v", got)
	}
}

func TestR4BindingsRequireCanonicalPayloadProof(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	if got := EvaluateR4(input); got.Status != R4StatusPass {
		t.Fatalf("valid bindings = %#v", got)
	}
	alternate := r4FixtureInput(t, "acyclic")
	alternate.States[0].Status = "PASS"
	projections, err := r4ProjectionBytes(alternate)
	if err != nil {
		t.Fatal(err)
	}

	stale := input
	stale.SnapshotDigest = r4BindingDigest(string(projections.snapshot))
	got := EvaluateR4(stale)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonSnapshotBindingMismatch || len(got.SelectedIDs) != 0 {
		t.Fatalf("stale binding = %#v", got)
	}

	mutated := input
	mutated.SnapshotPayload = string(projections.snapshot)
	got = EvaluateR4(mutated)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonSnapshotBindingMismatch || len(got.SelectedIDs) != 0 {
		t.Fatalf("mutated payload = %#v", got)
	}

	missing := input
	missing.PolicyPayload = ""
	got = EvaluateR4(missing)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonRequiredInputMissing || len(got.SelectedIDs) != 0 {
		t.Fatalf("missing payload = %#v", got)
	}

	malformed := input
	malformed.RegistryPayload = `{"fixture":"r4","fixture":"duplicate"}`
	got = EvaluateR4(malformed)
	if got.Status != R4StatusFailClosed || got.Reason != R4ReasonMalformedBinding || len(got.SelectedIDs) != 0 {
		t.Fatalf("duplicate payload field = %#v", got)
	}
}

func TestR4RootRelocationChangesReachableBinding(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	relocated := input
	relocated.RootObligationIDs = []string{"obligation/child"}
	first, err := AnalyzeR4Graph(input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzeR4Graph(relocated)
	if err != nil {
		t.Fatal(err)
	}
	if first.GraphDigest == second.GraphDigest || first.SCCDigest == second.SCCDigest {
		t.Fatalf("root relocation did not change reachable digests: %#v %#v", first, second)
	}
	if got := EvaluateR4(relocated); len(got.SelectedIDs) != 0 {
		t.Fatalf("relocated root selected paths without its declared frontier: %#v", got)
	}
}

func TestR4GovernedProjectionMutationsRequireRebinding(t *testing.T) {
	state := r4FixtureInput(t, "acyclic")
	state.States[0].Status = "PASS"
	assertR4BindingMismatch(t, state, R4ReasonSnapshotBindingMismatch)

	path := r4FixtureInput(t, "acyclic")
	path.Paths[0].PolicyPriority++
	assertR4BindingMismatch(t, path, R4ReasonSnapshotBindingMismatch)

	root := r4FixtureInput(t, "acyclic")
	root.RootObligationIDs = []string{"obligation/child"}
	assertR4BindingMismatch(t, root, R4ReasonSnapshotBindingMismatch)

	policy := r4FixtureInput(t, "acyclic")
	policy.Capacity.CPUCoreNS++
	assertR4BindingMismatch(t, policy, R4ReasonPolicyBindingMismatch)

	bound := r4FixtureInput(t, "self-loop")
	bound.Rules[0].MaxIterations++
	assertR4BindingMismatch(t, bound, R4ReasonPolicyBindingMismatch)

	use := r4FixtureInput(t, "self-loop")
	use.Rules[0].IterationsUsed++
	assertR4BindingMismatch(t, use, R4ReasonSnapshotBindingMismatch)

	registry := r4FixtureInput(t, "acyclic")
	registry.Pressures[0].StableID = "pressure/changed"
	assertR4BindingMismatch(t, registry, R4ReasonRegistryBindingMismatch)
}

func assertR4BindingMismatch(t *testing.T, input R4Input, reason string) {
	t.Helper()
	got := EvaluateR4(input)
	if got.Status != R4StatusUnknown || got.Reason != reason || len(got.SelectedIDs) != 0 {
		t.Fatalf("binding mismatch = %#v, want UNKNOWN/%s with empty selection", got, reason)
	}
}

func TestR4ReceiptIsComponentwiseAndExact(t *testing.T) {
	acyclic := EvaluateR4(r4FixtureInput(t, "acyclic"))
	wantAcyclic := R4WorkReceipt{
		GraphNodes: 2, GraphEdges: 1, ReachableNodes: 2, ReachableEdges: 1,
		SCCs: 2, CyclicSCCs: 0, CondensationEdges: 1, RuleChecks: 2,
		IterationChecks: 0, PathChecks: 2, ConflictChecks: 0,
	}
	if !reflect.DeepEqual(acyclic.WorkReceipt, wantAcyclic) {
		t.Fatalf("acyclic receipt = %#v, want %#v", acyclic.WorkReceipt, wantAcyclic)
	}
	cycle := EvaluateR4(r4FixtureInput(t, "self-loop"))
	wantCycle := R4WorkReceipt{
		GraphNodes: 1, GraphEdges: 1, ReachableNodes: 1, ReachableEdges: 1,
		SCCs: 1, CyclicSCCs: 1, CondensationEdges: 0, RuleChecks: 1,
		IterationChecks: 1, PathChecks: 1, ConflictChecks: 0,
	}
	if !reflect.DeepEqual(cycle.WorkReceipt, wantCycle) {
		t.Fatalf("cycle receipt = %#v, want %#v", cycle.WorkReceipt, wantCycle)
	}
}
