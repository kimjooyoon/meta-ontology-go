package query

import (
	"testing"
)

func TestEnvelopeTraversalQuotaStopsEdgeScanBeforeResultRows(t *testing.T) {
	graph := New()
	root := id("urn:query:quota:root")
	noise := id("urn:query:quota:noise")
	target := id("urn:query:quota:target")
	assertAdd(t, graph, NewFact(root, Used, noise))
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))

	limited := traversalEnvelope(root, LayerDeterministic, 1, 1)
	limited.Relation = WasDerivedFrom
	response, err := graph.Execute(limited)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 0 {
		t.Fatalf("edge quota scanned beyond its limit: %#v", response.Result)
	}

	full := limited
	full.Limit = 2
	response, err = graph.Execute(full)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 1 || response.Result.DeterministicPaths[0].Last() != target {
		t.Fatalf("larger traversal quota did not reach the matching edge: %#v", response.Result)
	}
}
