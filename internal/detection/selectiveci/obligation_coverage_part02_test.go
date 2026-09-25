package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"reflect"
	"testing"
)

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
