package bodycodegen

import (
	"context"
	"testing"
)

const scalarModelContextSource = `package contextdecision
namespace contextdecision
entity Integer id "urn:contextdecision:integer"
entity Boolean id "urn:contextdecision:boolean"
activity Charge(Integer) -> Integer id "urn:contextdecision:charge" computes "return 19 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 19를 빼기. Subtract nineteen from the input."
 criteria { predicate "relation" activity "Relation" input "[22]" holdout_input "[40]" }
 attempts "2"
}
activity Difference(Integer) -> Integer id "urn:contextdecision:difference" computes "return input - 19"
activity Relation(Integer,Integer) -> Boolean id "urn:contextdecision:relation" computes "return input1 == Difference(input0)"
`

func TestScalarSourceModelContextIncludesPredicateHelper(t *testing.T) {
	name := scalarBridgeSyntheticArtifact(t)
	r, err := ExportRecordAssemblyModelContext(context.Background(), "context-model.gooo",
		[]byte(scalarModelContextSource), "Charge", name, RecordModelContextOptions{ValueFlow: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.ModelPredictions != 0 || r.CandidateTests != 0 || r.ScalarFeatures == nil ||
		r.ScalarFeatures.ChoiceCount != 1 || r.ScalarFeatures.PredicateCount != 1 ||
		r.ModelCompatibility == nil || r.ModelCompatibility.Status != "READY_FOR_RANKING" ||
		r.ActivityID != "urn:contextdecision:charge" || r.Context.Status != "ENCODED" {
		t.Fatal("scalar source model preflight differs", r)
	}
	if r.CriteriaFlow == nil || len(r.CriteriaFlow.Predicates) != 1 ||
		len(r.CriteriaFlow.Predicates[0].ValueFlow.Helpers) != 1 || r.ExpandedPlan != nil {
		t.Fatal("predicate helper source or finite-input separation lost", r.CriteriaFlow)
	}
}
