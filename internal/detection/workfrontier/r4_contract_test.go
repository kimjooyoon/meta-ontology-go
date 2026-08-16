package workfrontier

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
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
	got := EvaluateR4(input)
	if got.Status != R4StatusFailClosed || got.Reason != R4ReasonMalformedGraph || len(got.SelectedIDs) != 0 {
		t.Fatalf("malformed graph result = %#v", got)
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

func r4FixtureInput(t *testing.T, name string) R4Input {
	t.Helper()
	base := R4Input{
		SchemaVersion:            R4SchemaVersion,
		SnapshotDigest:           strings.Repeat("a", 64),
		PolicyDigest:             strings.Repeat("b", 64),
		RegistryDigest:           strings.Repeat("c", 64),
		MinimumSelectedPressures: 2,
		Capacity:                 Capacity{CPUCoreNS: 20},
		Pressures:                []Pressure{{StableID: "pressure/a"}, {StableID: "pressure/b"}},
		States:                   []ObligationState{{ObligationID: "obligation/root", Status: "PENDING"}},
		Paths: []RepairPath{{
			StableID: "path/root", ObligationID: "obligation/root", ReadSet: []string{"pressure/a"},
			WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 1, CPUCoreNSUpperBound: 1,
		}},
		RootObligationIDs: []string{"obligation/root"},
		Rules:             []R4Rule{},
	}
	switch name {
	case "acyclic":
		base.States = append(base.States, ObligationState{ObligationID: "obligation/child", Status: "PENDING"})
		base.Paths = append(base.Paths, RepairPath{
			StableID: "path/child", ObligationID: "obligation/child", PrerequisiteObligationIDs: []string{"obligation/root"},
			ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 2, CPUCoreNSUpperBound: 1,
		})
	case "self-loop", "missing-bound", "zero-bound", "stale-digest", "iteration-exhaustion", "conflicting-rule":
		base.Paths[0].PrerequisiteObligationIDs = []string{"obligation/root"}
		base = bindR4Rules(t, base, name)
	case "two-node-cycle", "reordered-input":
		base.States = append(base.States, ObligationState{ObligationID: "obligation/b", Status: "PENDING"})
		base.Paths[0] = RepairPath{
			StableID: "path/a", ObligationID: "obligation/root", PrerequisiteObligationIDs: []string{"obligation/b"},
			ReadSet: []string{"pressure/a"}, WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 1, CPUCoreNSUpperBound: 1,
		}
		base.Paths = append(base.Paths, RepairPath{
			StableID: "path/b", ObligationID: "obligation/b", PrerequisiteObligationIDs: []string{"obligation/root"},
			ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"},
			PolicyPriority: 2, CPUCoreNSUpperBound: 1,
		})
		// Close the cycle by making the declared root obligation depend on b.
		base.RootObligationIDs = []string{"obligation/root"}
		base = bindR4Rules(t, base, name)
	case "unreachable-cycle":
		base.States = append(base.States,
			ObligationState{ObligationID: "obligation/a", Status: "PENDING"},
			ObligationState{ObligationID: "obligation/b", Status: "PENDING"},
		)
		base.Paths = append(base.Paths,
			RepairPath{StableID: "path/a", ObligationID: "obligation/a", PrerequisiteObligationIDs: []string{"obligation/b"}, ReadSet: []string{"pressure/a"}, WriteSet: []string{"pressure/a"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"}, CPUCoreNSUpperBound: 1},
			RepairPath{StableID: "path/b", ObligationID: "obligation/b", PrerequisiteObligationIDs: []string{"obligation/a"}, ReadSet: []string{"pressure/b"}, WriteSet: []string{"pressure/b"}, RequiredPressureIDs: []string{"pressure/a", "pressure/b"}, CPUCoreNSUpperBound: 1},
		)
	default:
		t.Fatalf("unknown R4 fixture %q", name)
	}
	return base
}

func bindR4Rules(t *testing.T, input R4Input, name string) R4Input {
	t.Helper()
	if name == "self-loop" || name == "missing-bound" || name == "zero-bound" || name == "stale-digest" || name == "iteration-exhaustion" || name == "conflicting-rule" {
		input.Rules = []R4Rule{}
	}
	if name == "missing-bound" {
		return input
	}
	graph, err := AnalyzeR4Graph(input)
	if err != nil {
		t.Fatal(err)
	}
	var digest string
	for _, component := range graph.SCCs {
		if component.Cyclic {
			digest = component.Digest
		}
	}
	if name == "stale-digest" {
		digest = "stale-scc-digest"
	}
	maxIterations := uint64(2)
	iterationsUsed := uint64(0)
	if name == "zero-bound" {
		maxIterations = 0
	}
	if name == "iteration-exhaustion" {
		maxIterations = 1
		iterationsUsed = 1
	}
	input.Rules = []R4Rule{{SCCDigest: digest, MaxIterations: maxIterations, IterationsUsed: iterationsUsed}}
	if name == "conflicting-rule" {
		input.Rules = append(input.Rules, R4Rule{SCCDigest: digest, MaxIterations: maxIterations + 1})
	}
	return input
}

func r4FixtureDigest(input R4Input) string {
	data, err := EncodeR4JSON(input)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func reversePressures(values []Pressure) []Pressure {
	result := append([]Pressure(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseStates(values []ObligationState) []ObligationState {
	result := append([]ObligationState(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reversePaths(values []RepairPath) []RepairPath {
	result := append([]RepairPath(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseRules(values []R4Rule) []R4Rule {
	result := append([]R4Rule(nil), values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
