package selectiveci

import (
	"math"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

func TestBaselineMultiRootCoverageIsLaunderedByCoveredRoot(t *testing.T) {
	input := completeInput()
	input.Base.Files = append(input.Base.Files, SnapshotFile{
		Path:        "billing/customer.gooo",
		BlobDigest:  digest("customer-base"),
		SemanticIDs: []string{"urn:selectiveci:entity/customer"},
	})
	input.Head.Files = append(input.Head.Files, SnapshotFile{
		Path:        "billing/customer.gooo",
		BlobDigest:  digest("customer-head"),
		SemanticIDs: []string{"urn:selectiveci:entity/customer"},
	})
	input.Base.Digest = input.Base.ComputedDigest()
	input.Head.Digest = input.Head.ComputedDigest()
	for index := range input.Receipts {
		input.Receipts[index].SnapshotDigest = input.Head.Digest
	}

	got := Plan(input)
	if got.Status != StatusFullSuiteFallback || got.ReasonCode != "MISSING_OBLIGATION" {
		t.Fatalf("multi-root status = %s/%s, want FULL_SUITE_FALLBACK/MISSING_OBLIGATION", got.Status, got.ReasonCode)
	}
	if len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs)+len(got.SelectedWorkIDs)+len(got.ResourceReceiptDigests)+len(got.ProvenancePathIDs) != 0 {
		t.Fatalf("uncovered root retained selection evidence: %#v", got)
	}
}

func TestBaselineZeroObligationSingleRootFallsBack(t *testing.T) {
	input := completeInput()
	input.Registry.Obligations = []ObligationBinding{}
	input.Registry.Nodes = nodesWithoutObligation(input.Registry.Nodes)
	input.Registry.Digest = input.Registry.ComputedDigest()

	got := Plan(input)
	if got.Status != StatusFullSuiteFallback || got.ReasonCode != "MISSING_OBLIGATION" {
		t.Fatalf("zero-obligation status = %s/%s, want FULL_SUITE_FALLBACK/MISSING_OBLIGATION", got.Status, got.ReasonCode)
	}
}

func TestPlanCoverageCommandFailuresClearSelectionEvidence(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		reason string
	}{
		{"missing command", func(input *Input) {
			input.Registry.Obligations[0].CommandIDs = []string{}
			input.Registry.Digest = input.Registry.ComputedDigest()
		}, ReasonMissingCommand},
		{"dangling command", func(input *Input) {
			input.Registry.Obligations[0].CommandIDs = []string{"urn:selectiveci:command/missing"}
			input.Registry.Digest = input.Registry.ComputedDigest()
		}, ReasonDanglingCommand},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := completeInput()
			test.mutate(&input)
			got := Plan(input)
			if got.Status != StatusFullSuiteFallback || got.ReasonCode != test.reason {
				t.Fatalf("plan = %s/%s, want FULL_SUITE_FALLBACK/%s", got.Status, got.ReasonCode, test.reason)
			}
			if len(got.SelectedCommandIDs)+len(got.SelectedGuardCommandIDs)+len(got.SelectedWorkIDs)+len(got.ResourceReceiptDigests)+len(got.ProvenancePathIDs) != 0 {
				t.Fatalf("plan retained selection evidence: %#v", got)
			}
		})
	}
}

func nodesWithoutObligation(nodes []impactgraph.Node) []impactgraph.Node {
	result := make([]impactgraph.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind != impactgraph.NodeKindObligation {
			result = append(result, node)
		}
	}
	return result
}

func TestObligationCoverageExactOneRootOneCommand(t *testing.T) {
	input := completeInput()
	coverageInput := coverageInputFromPlanInput(t, input, "urn:selectiveci:entity/order")
	got := ObserveObligationCoverage(coverageInput)
	if got.Decision != CoverageDecisionExact || got.Reason != CoverageReasonComplete {
		t.Fatalf("coverage = %s/%s, want EXACT/COMPLETE", got.Decision, got.Reason)
	}
	if got.ChangedRootCount != 1 || got.CoveredChangedRootCount != 1 || got.UncoveredChangedRootCount != 0 || got.RequiredObligationCount != 1 || got.BoundCommandCount != 1 || got.DeterministicWorkUnits != 3 {
		t.Fatalf("coverage counts = %#v", got)
	}
	if !reflect.DeepEqual(got.RequiredObligationIDs, []string{"urn:selectiveci:obligation/order"}) || len(got.UncoveredRootIDs) != 0 {
		t.Fatalf("coverage IDs = %#v", got)
	}
	if got.GraphDigest != coverageInput.Graph.Digest() || got.RegistryDigest != input.Registry.Digest || got.SnapshotDigest != input.Head.Digest || got.InputDigest != coverageInput.Digest() || got.OutputDigest != got.StableDigest() {
		t.Fatalf("coverage digests = %#v", got)
	}
}

func TestObligationCoverageZeroObligationAndTwoRoots(t *testing.T) {
	t.Run("zero obligation", func(t *testing.T) {
		input := completeInput()
		input.Registry.Obligations = []ObligationBinding{}
		input.Registry.Nodes = nodesWithoutObligation(input.Registry.Nodes)
		input.Registry.Digest = input.Registry.ComputedDigest()
		got := ObserveObligationCoverage(coverageInputFromPlanInput(t, input, "urn:selectiveci:entity/order"))
		assertMissingCoverage(t, got, CoverageReasonMissingObligation, 1, 0, 1)
	})
	t.Run("two roots one uncovered", func(t *testing.T) {
		input := completeInput()
		input.Base.Files = append(input.Base.Files, SnapshotFile{Path: "billing/customer.gooo", BlobDigest: digest("customer-base"), SemanticIDs: []string{"urn:selectiveci:entity/customer"}})
		input.Head.Files = append(input.Head.Files, SnapshotFile{Path: "billing/customer.gooo", BlobDigest: digest("customer-head"), SemanticIDs: []string{"urn:selectiveci:entity/customer"}})
		input.Base.Digest = input.Base.ComputedDigest()
		input.Head.Digest = input.Head.ComputedDigest()
		got := ObserveObligationCoverage(coverageInputFromPlanInput(t, input, "urn:selectiveci:entity/customer", "urn:selectiveci:entity/order"))
		assertMissingCoverage(t, got, CoverageReasonMissingObligation, 2, 1, 1)
		if got.RequiredObligationCount != 1 || got.BoundCommandCount != 1 || got.DeterministicWorkUnits != 4 {
			t.Fatalf("two-root counts = %#v", got)
		}
		if !reflect.DeepEqual(got.UncoveredRootIDs, []string{"urn:selectiveci:entity/customer"}) || len(got.RequiredObligationIDs) != 0 {
			t.Fatalf("two-root IDs = %#v", got)
		}
	})
}

func TestObligationCoverageOneRootTwoObligationsAndTransitiveDependency(t *testing.T) {
	t.Run("two obligations", func(t *testing.T) {
		input := completeInput()
		obligation := "urn:selectiveci:obligation/order-2"
		command := "urn:selectiveci:command/test-2"
		input.Registry.Nodes = append(input.Registry.Nodes, impactgraph.Node{ID: obligation, Kind: impactgraph.NodeKindObligation})
		input.Registry.Obligations = append(input.Registry.Obligations, ObligationBinding{ID: obligation, Subject: "urn:selectiveci:entity/order", CommandIDs: []string{command}})
		input.Registry.Commands = append(input.Registry.Commands, Command{ID: command, Argv: []string{"go", "test"}, WorkingDir: ".", CPUWorkUnits: 100, MemoryBytes: 1000})
		input.Registry.Digest = input.Registry.ComputedDigest()
		coverageInput := coverageInputFromPlanInput(t, input, "urn:selectiveci:entity/order")
		got := ObserveObligationCoverage(coverageInput)
		if got.Decision != CoverageDecisionExact || got.RequiredObligationCount != 2 || got.BoundCommandCount != 2 || got.DeterministicWorkUnits != 5 {
			t.Fatalf("two-obligation coverage = %#v", got)
		}
		coverageInput.Registry.Obligations[1].CommandIDs = []string{coverageInput.Registry.Obligations[0].CommandIDs[0]}
		coverageInput.Registry.Digest = coverageInput.Registry.ComputedDigest()
		coverageInput.Graph.RegistryDigest = coverageInput.Registry.Digest
		got = ObserveObligationCoverage(coverageInput)
		if got.Decision != CoverageDecisionExact || got.BoundCommandCount != 1 || got.DeterministicWorkUnits != 5 {
			t.Fatalf("shared-command coverage = %#v", got)
		}
	})
	t.Run("typed transitive dependency", func(t *testing.T) {
		input := typedTransitiveCoverageInput(t)
		got := ObserveObligationCoverage(input)
		if got.Decision != CoverageDecisionExact || got.Reason != CoverageReasonComplete || got.RequiredObligationCount != 1 || got.BoundCommandCount != 1 {
			t.Fatalf("transitive coverage = %#v", got)
		}
	})
}

func TestObligationCoverageCommandAndInputReasons(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*ObligationCoverageInput)
		reason CoverageReason
	}{
		{"zero command", func(input *ObligationCoverageInput) {
			input.Registry.Obligations[0].CommandIDs = []string{}
			input.Registry.Digest = input.Registry.ComputedDigest()
			input.Graph.RegistryDigest = input.Registry.Digest
		}, CoverageReasonMissingCommand},
		{"dangling command", func(input *ObligationCoverageInput) {
			input.Registry.Obligations[0].CommandIDs = []string{"urn:selectiveci:command/missing"}
			input.Registry.Digest = input.Registry.ComputedDigest()
			input.Graph.RegistryDigest = input.Registry.Digest
		}, CoverageReasonDanglingCommand},
		{"unknown root", func(input *ObligationCoverageInput) {
			input.ChangedRootIDs = []string{"urn:selectiveci:entity/missing"}
		}, CoverageReasonUnknownRoot},
		{"unsupported schema", func(input *ObligationCoverageInput) {
			input.SchemaVersion = "gooo/selective-ci-obligation-coverage/v0"
		}, CoverageReasonUnsupportedSchema},
		{"duplicate root", func(input *ObligationCoverageInput) {
			input.ChangedRootIDs = append(input.ChangedRootIDs, input.ChangedRootIDs[0])
		}, CoverageReasonDuplicateRoot},
		{"stale graph", func(input *ObligationCoverageInput) { input.Graph.RegistryDigest = digest("stale-graph") }, CoverageReasonStaleGraph},
		{"stale registry", func(input *ObligationCoverageInput) {
			input.Registry.Digest = digest("stale-registry")
		}, CoverageReasonStaleRegistry},
		{"stale snapshot", func(input *ObligationCoverageInput) { input.SnapshotDigest = digest("stale-snapshot") }, CoverageReasonStaleSnapshot},
		{"invalid graph", func(input *ObligationCoverageInput) { input.Graph.Version = "gooo/impact-graph/v0" }, CoverageReasonInvalidGraph},
		{"missing snapshot", func(input *ObligationCoverageInput) { input.SnapshotDigest = "" }, CoverageReasonInvalidSnapshot},
		{"missing roots", func(input *ObligationCoverageInput) { input.ChangedRootIDs = nil }, CoverageReasonMissingInput},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := coverageInputFromPlanInput(t, completeInput(), "urn:selectiveci:entity/order")
			test.mutate(&input)
			got := ObserveObligationCoverage(input)
			if got.Decision != CoverageDecisionUnknown || got.Reason != test.reason || len(got.RequiredObligationIDs) != 0 || !got.FullSuiteRequired {
				t.Fatalf("coverage = %#v, want UNKNOWN/%s without required IDs", got, test.reason)
			}
		})
	}
}

func TestObligationCoverageNoChangePermutationAndCanonicalOutput(t *testing.T) {
	input := coverageInputFromPlanInput(t, completeInput(), []string{}...)
	got := ObserveObligationCoverage(input)
	if got.Decision != CoverageDecisionExact || got.Reason != CoverageReasonNoChange || got.DeterministicWorkUnits != 0 {
		t.Fatalf("no-change coverage = %#v", got)
	}
	permuted := input
	permuted.ChangedRootIDs = []string{}
	permuted.Graph.Nodes = reverseImpactNodes(permuted.Graph.Nodes)
	permuted.Graph.Edges = reverseImpactEdges(permuted.Graph.Edges)
	permuted.Registry.Obligations = reverseObligationsForCoverage(permuted.Registry.Obligations)
	permuted.Registry.Commands = reverseCommandsForCoverage(permuted.Registry.Commands)
	permuted.Registry.Digest = permuted.Registry.ComputedDigest()
	permuted.Graph.RegistryDigest = permuted.Registry.Digest
	left, err := EncodeObligationCoverageJSON(got)
	if err != nil {
		t.Fatal(err)
	}
	right, err := EncodeObligationCoverageJSON(ObserveObligationCoverage(permuted))
	if err != nil {
		t.Fatal(err)
	}
	if string(left) != string(right) {
		t.Fatalf("permutation changed coverage output:\n%s\n%s", left, right)
	}
	inputJSON, err := EncodeObligationCoverageInputJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeObligationCoverageInputJSON(inputJSON); err != nil {
		t.Fatalf("input JSON round trip failed: %v", err)
	}
	if _, err := DecodeObligationCoverageInputJSON([]byte(`{"schema_version":"gooo/selective-ci-obligation-coverage/v1","schema_version":"gooo/selective-ci-obligation-coverage/v1"}`)); err == nil {
		t.Fatal("duplicate input fields were accepted")
	}
}

func TestObligationCoverageWorkOverflowAndReasonLabels(t *testing.T) {
	if _, ok := coverageWorkUnits(math.MaxUint64, 1, 0); ok {
		t.Fatal("root work overflow was accepted")
	}
	if _, ok := coverageWorkUnits(1, math.MaxUint64, 1); ok {
		t.Fatal("obligation work overflow was accepted")
	}
	if _, ok := coverageWorkUnits(1, 1, math.MaxUint64); ok {
		t.Fatal("command-binding work overflow was accepted")
	}
	if _, err := DecodeObligationCoverageJSON([]byte(`{"schema_version":"gooo/selective-ci-obligation-coverage/v1","decision":"UNKNOWN","reason":"not-a-reason","full_suite_required":true,"changed_root_count":0,"covered_changed_root_count":0,"uncovered_changed_root_count":0,"required_obligation_count":0,"bound_command_count":0,"deterministic_work_units":0,"uncovered_root_ids":[],"required_obligation_ids":[],"graph_digest":"","registry_digest":"","snapshot_digest":"","input_digest":"","output_digest":""}`)); err == nil {
		t.Fatal("unknown reason label was accepted")
	}
}

func assertMissingCoverage(t *testing.T, got ObligationCoverageResult, reason CoverageReason, roots, covered, uncovered uint64) {
	t.Helper()
	if got.Decision != CoverageDecisionUnknown || got.Reason != reason || !got.FullSuiteRequired {
		t.Fatalf("coverage = %#v, want UNKNOWN/%s", got, reason)
	}
	if got.ChangedRootCount != roots || got.CoveredChangedRootCount != covered || got.UncoveredChangedRootCount != uncovered {
		t.Fatalf("root counts = %#v", got)
	}
	if len(got.RequiredObligationIDs) != 0 {
		t.Fatalf("unknown coverage exposed required IDs: %v", got.RequiredObligationIDs)
	}
}

func coverageInputFromPlanInput(t *testing.T, input Input, roots ...string) ObligationCoverageInput {
	t.Helper()
	graph, err := buildGraph(input)
	if err != nil {
		t.Fatalf("build coverage graph: %v", err)
	}
	return ObligationCoverageInput{SchemaVersion: ObligationCoverageSchemaVersion, Graph: graph, Registry: input.Registry, SnapshotDigest: input.Head.Digest, ChangedRootIDs: roots}
}

func typedTransitiveCoverageInput(t *testing.T) ObligationCoverageInput {
	t.Helper()
	root := "urn:selectiveci:entity/root"
	packageID := "urn:selectiveci:package/pkg"
	obligation := "urn:selectiveci:obligation/root"
	command := "urn:selectiveci:command/root"
	snapshot := digest("typed-snapshot")
	registry := Registry{SchemaVersion: RegistrySchemaVersion, PolicyDigest: digest("typed-policy"), Nodes: []impactgraph.Node{{ID: root, Kind: impactgraph.NodeKindSemantic}, {ID: packageID, Kind: impactgraph.NodeKindGoPackage}, {ID: obligation, Kind: impactgraph.NodeKindObligation}}, DependencyEdges: []DependencyEdge{{From: root, To: packageID, Kind: impactgraph.EdgeKindProjectsTo}, {From: packageID, To: obligation, Kind: impactgraph.EdgeKindAffects}}, Obligations: []ObligationBinding{{ID: obligation, Subject: packageID, CommandIDs: []string{command}}}, Commands: []Command{{ID: command, Argv: []string{"go", "test"}, WorkingDir: ".", CPUWorkUnits: 1, MemoryBytes: 1}}, GlobalGuardCommands: []Command{}}
	registry.Digest = registry.ComputedDigest()
	graph := impactgraph.Graph{Version: impactgraph.SchemaVersion, SnapshotDigest: snapshot, RegistryDigest: registry.Digest, PolicyDigest: registry.PolicyDigest, Nodes: registry.Nodes, Edges: []impactgraph.Edge{{From: root, To: packageID, Kind: impactgraph.EdgeKindProjectsTo}, {From: packageID, To: obligation, Kind: impactgraph.EdgeKindAffects}}}
	return ObligationCoverageInput{SchemaVersion: ObligationCoverageSchemaVersion, Graph: graph, Registry: registry, SnapshotDigest: snapshot, ChangedRootIDs: []string{root}}
}

func reverseImpactNodes(nodes []impactgraph.Node) []impactgraph.Node {
	result := append([]impactgraph.Node{}, nodes...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseImpactEdges(edges []impactgraph.Edge) []impactgraph.Edge {
	result := append([]impactgraph.Edge{}, edges...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseObligationsForCoverage(values []ObligationBinding) []ObligationBinding {
	result := append([]ObligationBinding{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}

func reverseCommandsForCoverage(values []Command) []Command {
	result := append([]Command{}, values...)
	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}
	return result
}
