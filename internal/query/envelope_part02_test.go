package query

import (
	"testing"
)

func TestEnvelopeFiltersLayersAndBoundsResults(t *testing.T) {
	graph := New()
	root, order := id("urn:query:activity:root"), id("urn:query:entity:order")
	invoice := id("urn:query:entity:invoice")
	assertAdd(t, graph, NewFact(root, Used, order))
	assertAdd(t, graph, NewCandidateFact(root, Used, invoice, "unresolved"))

	all := traversalEnvelope(root, LayerAll, 1, 1)
	response, err := graph.Execute(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 1 || len(response.Result.CandidatePaths) != 0 {
		t.Fatalf("limit did not prefer deterministic paths: %#v", response.Result)
	}
	all.Limit = 2
	response, err = graph.Execute(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 1 || len(response.Result.CandidatePaths) != 1 {
		t.Fatalf("candidate paths did not follow deterministic paths: %#v", response.Result)
	}

	candidate := traversalEnvelope(root, LayerCandidate, 1, 10)
	response, err = graph.Execute(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 0 || len(response.Result.CandidatePaths) != 1 {
		t.Fatalf("candidate layer leaked another layer: %#v", response.Result)
	}

	exact := exactEnvelope(root, invoice, LayerCandidate)
	response, err = graph.Execute(exact)
	if err != nil || len(response.Result.CandidateMatches) != 1 || len(response.Result.DeterministicMatches) != 0 {
		t.Fatalf("candidate exact result = %#v, err=%v", response.Result, err)
	}
}
