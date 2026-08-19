package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestInferenceProjectionCompletenessAndTypedSelectors(t *testing.T) {
	path, edges := inferenceQueryFixture(t)
	before := path
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(path, before) {
		t.Fatal("projection constructor mutated the path")
	}
	request := inferenceQueryRequest()
	request.IncludeClaims = true
	request.IncludeEvidence = true
	result, err := projection.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Complete || len(result.Edges) != 6 || len(result.Claims) != 2 || len(result.Evidence) != 8 {
		t.Fatalf("complete typed result = %#v", result)
	}
	byRecord := make(map[ID]semantic.InferenceEdge, len(edges))
	for _, edge := range edges {
		byRecord[ID(edge.RecordID.String())] = edge
	}
	for _, row := range result.Edges {
		edge := byRecord[row.RecordID]
		if edge.RecordID == "" || row.SubjectID != ID(edge.SubjectID.String()) ||
			row.ObjectID != ID(edge.ObjectID.String()) || row.Kind != edge.Kind ||
			row.Phase != edge.Phase.Phase || row.AuthorityLayer != edge.Authority.Layer ||
			row.AuthorityEffect != edge.Authority.Effect || len(row.Evidence) != 1 {
			t.Fatalf("edge lost typed identity: %#v", row)
		}
	}
	for _, row := range result.Claims {
		if row.Kind != semantic.SemanticDelta && row.Kind != semantic.NoSemanticDelta {
			t.Fatalf("claim crossed closed semantic-change sum: %#v", row)
		}
	}
	selectors := []InferenceQuery{
		{RecordID: result.Edges[0].RecordID},
		{SubjectID: result.Edges[0].SubjectID},
		{ObjectID: result.Edges[0].ObjectID},
		{Phase: result.Edges[0].Phase},
		{Layer: result.Edges[0].AuthorityLayer},
		{Effect: result.Edges[0].AuthorityEffect},
	}
	for _, selector := range selectors {
		request := inferenceQueryRequest()
		request.RecordID, request.SubjectID, request.ObjectID = selector.RecordID, selector.SubjectID, selector.ObjectID
		request.Phase, request.Layer, request.Effect = selector.Phase, selector.Layer, selector.Effect
		selected, selectorErr := projection.Execute(request)
		if selectorErr != nil || !selected.Complete || len(selected.Edges) == 0 {
			t.Fatalf("typed selector %#v = %#v err=%v", selector, selected, selectorErr)
		}
	}
	byKind := inferenceQueryRequest()
	byKind.Kind = semantic.InferenceAcceptedLift
	byKind.EvidenceID = ID(edges[4].Evidence[0].ID.String())
	byKind.IncludeEvidence = true
	byKind.Limit = 4
	lift, err := projection.Query(byKind)
	if err != nil || !lift.Complete || len(lift.Edges) != 1 || len(lift.Evidence) != 1 {
		t.Fatalf("accepted lift selector = %#v err=%v", lift, err)
	}
	if lift.Edges[0].AcceptanceReceipt != byKind.EvidenceID || !lift.Evidence[0].SourceBacked {
		t.Fatalf("accepted lift receipt = %#v", lift)
	}
	claimQuery := inferenceQueryRequest()
	claimQuery.ClaimKind = semantic.SemanticDelta
	claimQuery.IncludeClaims = true
	claimResult, err := projection.Execute(claimQuery)
	if err != nil || len(claimResult.Claims) != 1 || len(claimResult.Edges) != 0 || claimResult.Claims[0].Kind != semantic.SemanticDelta {
		t.Fatalf("semantic claim separation = %#v err=%v", claimResult, err)
	}
}
