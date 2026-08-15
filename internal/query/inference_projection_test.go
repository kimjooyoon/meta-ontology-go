package query

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestInferenceProjectionChainPermutationAndDigestReplay(t *testing.T) {
	path, _ := inferenceQueryFixture(t)
	permuted := path
	permuted.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	permuted.Claims = append([]semantic.SemanticChangeClaim(nil), path.Claims...)
	permuted.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	for left, right := 0, len(permuted.Edges)-1; left < right; left, right = left+1, right-1 {
		permuted.Edges[left], permuted.Edges[right] = permuted.Edges[right], permuted.Edges[left]
	}
	for left, right := 0, len(permuted.Evidence)-1; left < right; left, right = left+1, right-1 {
		permuted.Evidence[left], permuted.Evidence[right] = permuted.Evidence[right], permuted.Evidence[left]
	}
	leftProjection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	rightProjection, err := NewInferenceProjection(permuted)
	if err != nil {
		t.Fatal(err)
	}
	request := inferenceQueryRequest()
	request.Explain = true
	request.ChainStartID = ID(path.Edges[0].SubjectID.String())
	request.ChainEndID = ID(path.Edges[len(path.Edges)-1].ObjectID.String())
	leftResult, err := leftProjection.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	requestDigest, err := request.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	rightResult, err := rightProjection.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	if !leftResult.Complete || leftResult.Chain == nil || leftResult.Chain.Depth != 6 ||
		leftResult.CanonicalDigestValue() != rightResult.CanonicalDigestValue() ||
		leftResult.RequestHash != requestDigest || leftResult.Hash != leftResult.CanonicalDigestValue() {
		t.Fatalf("chain replay = %#v / %#v", leftResult, rightResult)
	}
	if leftResult.Chain.Edges[0].SubjectID != ID(path.Edges[0].SubjectID.String()) ||
		leftResult.Chain.Edges[5].ObjectID != ID(path.Edges[5].ObjectID.String()) {
		t.Fatalf("chain ordering = %#v", leftResult.Chain)
	}
}

func TestInferenceProjectionBoundariesNeverPassPartials(t *testing.T) {
	path, _ := inferenceQueryFixture(t)
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	exact := inferenceQueryRequest()
	exact.Limit = 6
	exact.MaxWork = 6
	exactResult, err := projection.Execute(exact)
	if err != nil || !exactResult.Complete || exactResult.Work.Used != 6 {
		t.Fatalf("exact work boundary = %#v err=%v", exactResult, err)
	}
	oneOver := exact
	oneOver.MaxWork = 5
	overResult, err := projection.Execute(oneOver)
	if !errors.Is(err, ErrInferenceQueryBudget) || overResult.Complete || len(overResult.Edges) != 0 || overResult.Work.Used != 5 {
		t.Fatalf("one-over work boundary = %#v err=%v", overResult, err)
	}
	rowOver := exact
	rowOver.MaxWork = 32
	rowOver.Limit = 5
	rowResult, err := projection.Execute(rowOver)
	if !errors.Is(err, ErrInferenceQueryBudget) || rowResult.Complete || len(rowResult.Edges) != 0 {
		t.Fatalf("row overrun = %#v err=%v", rowResult, err)
	}
	unsupported := exact
	unsupported.Predicate = "not-a-semantic-predicate"
	badPredicate, err := projection.Execute(unsupported)
	if !errors.Is(err, ErrInferenceUnsupportedPred) || badPredicate.Complete {
		t.Fatalf("unsupported predicate = %#v err=%v", badPredicate, err)
	}
	depth := exact
	depth.Explain = true
	depth.MaxDepth = 5
	depthResult, err := projection.Execute(depth)
	if !errors.Is(err, ErrInferenceQueryBudget) || depthResult.Complete || len(depthResult.Edges) != 0 {
		t.Fatalf("depth overrun = %#v err=%v", depthResult, err)
	}
}

func TestInferenceProjectionRejectsInvalidAuthorityAndChains(t *testing.T) {
	path, edges := inferenceQueryFixture(t)
	lift := path
	lift.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	lift.Edges[4].AcceptanceReceipt = ""
	invalidResult, err := QueryInferencePath(lift, inferenceQueryRequest())
	if err == nil || invalidResult.Complete || invalidResult.Status != ResponseError || len(invalidResult.Edges) != 0 {
		t.Fatalf("invalid lift result = %#v err=%v", invalidResult, err)
	}
	stale := path
	stale.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	stale.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	stale.Evidence[0].After.Semantic = inferenceQueryDigest("stale")
	staleResult, err := QueryInferencePath(stale, inferenceQueryRequest())
	if err == nil || staleResult.Complete || !errors.Is(err, semantic.ErrInferencePath) {
		t.Fatalf("stale path result = %#v err=%v", staleResult, err)
	}
	ambiguous := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	first := edges[0]
	second := edges[1]
	second.SubjectID = first.SubjectID
	for _, edge := range []semantic.InferenceEdge{first, second} {
		evidenceID := edge.Evidence[0].ID
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls}
		ambiguous.Edges = append(ambiguous.Edges, edge)
		ambiguous.Evidence = append(ambiguous.Evidence, evidence)
	}
	ambiguousRequest := inferenceQueryRequest()
	ambiguousRequest.Explain = true
	ambiguousResult, err := QueryInferencePath(ambiguous, ambiguousRequest)
	if err == nil || ambiguousResult.Complete || !errors.Is(err, ErrInferenceChain) || len(ambiguousResult.Edges) != 0 {
		t.Fatalf("ambiguous chain result = %#v err=%v", ambiguousResult, err)
	}
	cycle := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	cycleA := edges[0]
	cycleB := edges[1]
	cycleA.SubjectID, cycleA.ObjectID = inferenceQueryID("cycle/a"), inferenceQueryID("cycle/b")
	cycleB.SubjectID, cycleB.ObjectID = inferenceQueryID("cycle/b"), inferenceQueryID("cycle/a")
	for _, edge := range []semantic.InferenceEdge{cycleA, cycleB} {
		evidenceID := edge.Evidence[0].ID
		cycle.Edges = append(cycle.Edges, edge)
		cycle.Evidence = append(cycle.Evidence, semantic.InferenceEvidence{ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls})
	}
	cycleRequest := inferenceQueryRequest()
	cycleRequest.Explain = true
	cycleResult, err := QueryInferencePath(cycle, cycleRequest)
	if err == nil || cycleResult.Complete || !errors.Is(err, ErrInferenceChain) {
		t.Fatalf("cycle chain result = %#v err=%v", cycleResult, err)
	}
}

func projectionForPath(t *testing.T, path semantic.InferencePathV1) *InferenceProjection {
	t.Helper()
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	return projection
}
