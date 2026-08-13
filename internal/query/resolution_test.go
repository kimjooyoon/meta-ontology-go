package query

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestGeneratedResolutionViewFromSemanticIR(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	business := mustSemanticEntity(t, "billing://entity/order", "Order")
	activity := mustSemanticActivity(t, "billing://activity/pay", "PayOrder")
	generated := mustSemanticEntity(t, "billing://entity/payment-code", "PaymentCode")
	for _, node := range []semantic.Node{business, activity, generated} {
		if err := ir.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := ir.AddFact(semantic.NewUsedFact(activity.ID, business.ID)); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddFact(semantic.NewWasGeneratedByFact(generated.ID, activity.ID)); err != nil {
		t.Fatal(err)
	}
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	beforeCanonical, beforeNodes, beforeHash := graph.Canonical(), graph.Nodes(), graph.StableHash()
	request := resolutionRequest(ID(business.ID.String()), LayerDeterministic, 10)
	response, err := graph.ResolveGeneratedCode(request)
	if err != nil || response.Status != ResolutionResolved {
		t.Fatalf("resolution failed: %#v %v", response, err)
	}
	if len(response.Deterministic) != 1 || len(response.Candidates) != 0 {
		t.Fatalf("resolution layers = %#v/%#v", response.Deterministic, response.Candidates)
	}
	row := response.Deterministic[0]
	if row.Business != ID(business.ID.String()) || row.Activity != ID(activity.ID.String()) ||
		row.GeneratedEntity != ID(generated.ID.String()) || row.Depth != ResolutionMaxDepth {
		t.Fatalf("resolution row = %#v", row)
	}
	if row.Status != DerivedFactStatus || row.SourceLayer != FactDeterministic.String() {
		t.Fatalf("resolution authority = %#v", row)
	}
	if response.Metadata.SemanticDigest != ir.StableHash() ||
		response.Metadata.DerivedStatus != DerivedStatusNonAuthoritative {
		t.Fatalf("resolution metadata = %#v", response.Metadata)
	}
	if label := resolutionLabel(response.Metadata); label.Authority != "derived" || label.Status != ResolutionResolved {
		t.Fatalf("resolution label = %#v", label)
	}
	if label := resolutionLabel(response.Metadata); label.View != "resolution_view" {
		t.Fatalf("resolution label view = %#v", label)
	}
	if response.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("resolution fabricated provenance = %#v", response.Metadata)
	}
	if label := authorityLabel(response.Metadata, "generated_go"); label.Status != "unavailable" {
		t.Fatalf("resolution fabricated generated Go = %#v", label)
	}
	requestHash, err := request.CanonicalDigest()
	if err != nil || response.RequestHash != requestHash {
		t.Fatalf("request digest = %q/%q", response.RequestHash, requestHash)
	}
	responseHash, err := response.CanonicalDigest()
	if err != nil || response.Hash != responseHash {
		t.Fatalf("response digest = %q/%q", response.Hash, responseHash)
	}
	if graph.Canonical() != beforeCanonical || !reflect.DeepEqual(graph.Nodes(), beforeNodes) ||
		graph.StableHash() != beforeHash {
		t.Fatal("resolution mutated graph authority")
	}
}

func TestGeneratedResolutionPermutationAndCandidateSeparation(t *testing.T) {
	business := id("urn:resolution:entity:business")
	activityDet := id("urn:resolution:activity:det")
	activityCandidate := id("urn:resolution:activity:candidate")
	generatedDet := id("urn:resolution:entity:generated-det")
	generatedCandidate := id("urn:resolution:entity:generated-candidate")
	facts := []Fact{
		NewFact(activityDet, Used, business),
		NewFact(generatedDet, WasGeneratedBy, activityDet),
		NewCandidateFact(activityCandidate, Used, business, "ambiguous activity"),
		NewCandidateFact(generatedCandidate, WasGeneratedBy, activityCandidate, "ambiguous output"),
	}
	first := newResolutionGraph(t, facts)
	secondFacts := append([]Fact(nil), facts...)
	for left, right := 0, len(secondFacts)-1; left < right; left, right = left+1, right-1 {
		secondFacts[left], secondFacts[right] = secondFacts[right], secondFacts[left]
	}
	second := newResolutionGraph(t, secondFacts)
	request := resolutionRequest(business, LayerAll, 10)
	want, err := first.ResolveGeneratedCode(request)
	if err != nil {
		t.Fatal(err)
	}
	got, err := second.ResolveGeneratedCode(request)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := want.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	gotJSON, err := got.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantJSON, gotJSON) || want.Hash != got.Hash {
		t.Fatalf("permutation changed resolution: %q/%q", want.Hash, got.Hash)
	}
	if len(want.Deterministic) != 1 || len(want.Candidates) != 1 ||
		want.Candidates[0].SourceLayer != FactCandidate.String() {
		t.Fatalf("resolution layers leaked: %#v/%#v", want.Deterministic, want.Candidates)
	}
	candidate, err := first.ResolveGeneratedCode(resolutionRequest(business, LayerCandidate, 10))
	if err != nil || len(candidate.Deterministic) != 0 || len(candidate.Candidates) != 1 {
		t.Fatalf("candidate resolution = %#v %v", candidate, err)
	}
	limited, err := first.ResolveGeneratedCode(resolutionRequest(business, LayerAll, 1))
	if err != nil || len(limited.Deterministic) != 1 || len(limited.Candidates) != 0 {
		t.Fatalf("resolution limit did not prefer deterministic: %#v %v", limited, err)
	}
}

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

func TestGeneratedResolutionLimitBoundsCartesianExpansionAndReplays(t *testing.T) {
	business := id("urn:resolution:bounded:business")
	graph := New()
	if err := graph.AddNode(Node{ID: business, Kind: EntityNodeKind}); err != nil {
		t.Fatal(err)
	}
	for activityIndex := 0; activityIndex < 24; activityIndex++ {
		activity := id(fmt.Sprintf("urn:resolution:bounded:activity:%02d", activityIndex))
		if err := graph.AddNode(Node{ID: activity, Kind: ActivityNodeKind}); err != nil {
			t.Fatal(err)
		}
		assertAdd(t, graph, NewFact(activity, Used, business))
		for entityIndex := 0; entityIndex < 24; entityIndex++ {
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
	for run := 0; run < 3; run++ {
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

func newResolutionGraph(t *testing.T, facts []Fact) *Graph {
	t.Helper()
	graph := New()
	nodes := []Node{
		{ID: id("urn:resolution:entity:business"), Kind: EntityNodeKind},
		{ID: id("urn:resolution:activity:det"), Kind: ActivityNodeKind},
		{ID: id("urn:resolution:activity:candidate"), Kind: ActivityNodeKind},
		{ID: id("urn:resolution:entity:generated-det"), Kind: EntityNodeKind},
		{ID: id("urn:resolution:entity:generated-candidate"), Kind: EntityNodeKind},
	}
	for _, node := range nodes {
		if err := graph.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	for _, fact := range facts {
		assertAdd(t, graph, fact)
	}
	return graph
}

func resolutionLabel(metadata EnvelopeMetadata) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == "resolution_view" {
			return label
		}
	}
	return AuthorityLabel{}
}

func mustSemanticEntity(t *testing.T, raw, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewEntity(semantic.MustIdentity(raw), "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}

func mustSemanticActivity(t *testing.T, raw, name string) semantic.Node {
	t.Helper()
	node, err := semantic.NewActivity(semantic.MustIdentity(raw), "billing", name)
	if err != nil {
		t.Fatal(err)
	}
	return node
}
