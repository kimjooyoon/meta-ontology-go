package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func mutateR4Record(input *pathclosure.R4Input, id semantic.ID, mutate func(*pathclosure.R4Record)) {
	for index := range input.Records {
		if input.Records[index].ID == id {
			mutate(&input.Records[index])
		}
	}
}
func mutateR4Receipt(input *pathclosure.R4Input, id semantic.ID, mutate func(*pathclosure.R4Receipt)) {
	for index := range input.Receipts {
		if input.Receipts[index].ID == id {
			mutate(&input.Receipts[index])
		}
	}
}
func TestEvaluateR4CompleteFiniteBoundaryNeverAuthorizesPromotion(t *testing.T) {
	fixture := completeR4Fixture()
	got := pathclosure.EvaluateR4(fixture.input)
	if got.Status != pathclosure.PASS || got.Code != pathclosure.CodeR4ProofValid || !got.ProofValid || got.PromotionAuthorized {
		t.Fatalf("complete R4 result = %#v", got)
	}
	if !reflect.DeepEqual(got.CoveredPathIDs, []semantic.ID{fixture.path.ID}) || got.Cost != 5 {
		t.Fatalf("coverage/cost = %#v", got)
	}
}
