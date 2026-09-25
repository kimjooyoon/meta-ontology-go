package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
