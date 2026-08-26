package query

import (
	"errors"
	"testing"
)

func TestGeneratedResolutionDefersMissingAndRejectsInvalidEndpoints(t *testing.T) {
	business := id("urn:resolution:missing:business")
	graph := New()
	if err := graph.AddNode(Node{ID: business, Kind: EntityNodeKind}); err != nil {
		t.Fatal(err)
	}
	missing, err := graph.ResolveGeneratedCode(resolutionRequest(business, LayerAll, 10))
	if err != nil || missing.Status != StatusDeferred || len(missing.Deterministic)+len(missing.Candidates) != 0 {
		t.Fatalf("missing path was not deferred: %#v %v", missing, err)
	}
	unknown, err := graph.ResolveGeneratedCode(resolutionRequest(id("urn:resolution:missing:unknown"), LayerAll, 10))
	if !errors.Is(err, ErrUnknownEndpoint) || unknown.Error == nil || unknown.Error.Code != "unknown_endpoint" {
		t.Fatalf("unknown business was not rejected: %#v %v", unknown, err)
	}
	activity := id("urn:resolution:invalid:activity")
	if err := graph.AddNode(Node{ID: activity, Kind: ActivityNodeKind}); err != nil {
		t.Fatal(err)
	}
	invalid, err := graph.ResolveGeneratedCode(resolutionRequest(activity, LayerAll, 10))
	if !errors.Is(err, ErrInvalidResolution) || invalid.Error == nil || invalid.Error.Code != "invalid_business_kind" {
		t.Fatalf("invalid business kind was not rejected: %#v %v", invalid, err)
	}
	for _, depth := range []int{1, ResolutionMaxDepth + 1} {
		request := resolutionRequest(business, LayerAll, 10)
		request.MaxDepth = depth
		response, err := graph.ResolveGeneratedCode(request)
		if err == nil || response.Error == nil || response.Error.Code != "invalid_resolution_depth" {
			t.Fatalf("depth %d was not rejected: %#v %v", depth, response, err)
		}
	}
}
