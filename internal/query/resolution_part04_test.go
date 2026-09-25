package query

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGeneratedResolutionLimitBoundsCartesianExpansionAndReplays(t *testing.T) {
	business := id("urn:resolution:bounded:business")
	graph := New()
	if err := graph.AddNode(Node{ID: business, Kind: EntityNodeKind}); err != nil {
		t.Fatal(err)
	}
	for activityIndex := range 24 {
		activity := id(fmt.Sprintf("urn:resolution:bounded:activity:%02d", activityIndex))
		if err := graph.AddNode(Node{ID: activity, Kind: ActivityNodeKind}); err != nil {
			t.Fatal(err)
		}
		assertAdd(t, graph, NewFact(activity, Used, business))
		for entityIndex := range 24 {
			rawGenerated := fmt.Sprintf(
				"urn:resolution:bounded:generated:%02d:%02d", activityIndex, entityIndex,
			)
			generated := id(rawGenerated)
			if err := graph.AddNode(Node{ID: generated, Kind: EntityNodeKind}); err != nil {
				t.Fatal(err)
			}
			assertAdd(t, graph, NewFact(generated, WasGeneratedBy, activity))
		}
	}

	beforeCanonical, beforeHash, beforeNodes := graph.Canonical(), graph.StableHash(), graph.Nodes()
	limitedRequest := resolutionRequest(business, LayerDeterministic, 1)
	limited, err := graph.ResolveGeneratedCode(limitedRequest)
	if err != nil || len(limited.Deterministic) != 1 || len(limited.Candidates) != 0 {
		t.Fatalf("bounded resolution = %#v, err=%v", limited, err)
	}
	full, err := graph.ResolveGeneratedCode(resolutionRequest(business, LayerDeterministic, 10))
	if err != nil || len(full.Deterministic) != 10 {
		t.Fatalf("expanded resolution = %#v, err=%v", full, err)
	}
	if !reflect.DeepEqual(limited.Deterministic, full.Deterministic[:1]) {
		t.Fatalf("bounded resolution changed canonical prefix: %#v vs %#v", limited.Deterministic, full.Deterministic[:1])
	}
	for run := range 3 {
		replay, replayErr := graph.ResolveGeneratedCode(limitedRequest)
		if replayErr != nil || replay.Hash != limited.Hash {
			t.Fatalf("resolution replay %d changed: %#v, err=%v", run, replay, replayErr)
		}
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash ||
		!reflect.DeepEqual(graph.Nodes(), beforeNodes) {
		t.Fatal("bounded resolution mutated graph authority")
	}
}
func resolutionRequest(business ID, layer Layer, limit int) ResolutionRequest {
	return ResolutionRequest{
		Schema: ResolutionSchema, Business: business, Layer: layer,
		MaxDepth: ResolutionMaxDepth, Limit: limit,
	}
}
