package bidir

import (
	"strings"
	"testing"
)

const scalarCriteriaFixture = `package feecriteria
namespace feecriteria
entity Integer id "fee-criteria://integer"
entity Boolean id "fee-criteria://boolean"
activity Fee(Integer) -> Integer id "fee-criteria://fee" computes "return 11 - input" assembling {
 choice "direction" operand_order at "0" intent "입력에서 11을 뺀다. Subtract eleven from the input."
 criteria {
  predicate "difference" activity "Difference"
  input "[11]"
  holdout_input "[20]"
 }
 observe {
  criteria_strategy "distinguish_then_check"
  criteria_input "[17]"
  max_candidates "2"
  max_rounds "1"
 }
 attempts "2"
}
activity Difference(Integer,Integer) -> Boolean computes "return input1 == input0 - 11"
`

func TestScalarCriteriaBXKeepsFunctionTypeAndInputMeaning(t *testing.T) {
	doc := criteriaDocument(t, scalarCriteriaFixture)
	support := EntityFieldsV4Support()
	if err := CheckGetPutWithEntityFieldsSupport(doc, support); err != nil {
		t.Fatal(err)
	}
	core, err := LowerDocumentWithEntityFieldsSupport(doc, support)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, node := range core.Graph.Nodes() {
		if node.Name == "Fee" {
			found = true
			if node.Assembly.Criteria.Inputs[0] != "[11]" || node.Assembly.Observation.CriteriaInputs[0] != "[17]" {
				t.Fatal(node)
			}
		}
	}
	if !found {
		t.Fatal("scalar activity missing")
	}
}
func TestScalarCriteriaBXEditChangesSemanticMeaning(t *testing.T) {
	support := EntityFieldsV4Support()
	doc := criteriaDocument(t, scalarCriteriaFixture)
	changed := criteriaDocument(t, strings.Replace(scalarCriteriaFixture, "[17]", "[19]", 1))
	model, err := GetWithEntityFieldsSupport(changed, support)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckPutGetWithEntityFieldsSupport(doc, model, support); err != nil {
		t.Fatal(err)
	}
	before, err := LowerDocumentWithEntityFieldsSupport(doc, support)
	if err != nil {
		t.Fatal(err)
	}
	after, err := LowerDocumentWithEntityFieldsSupport(changed, support)
	if err != nil || before.StableHash() == after.StableHash() {
		t.Fatal("scalar question edit lost meaning", err)
	}
}
