package syntax

import (
	"reflect"
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

func TestScalarCriteriaFormattingRetainsOriginalTypeAndObligations(t *testing.T) {
	file, d := ParseFile("scalar.gooo", scalarCriteriaFixture)
	if d.HasErrors() {
		t.Fatal(d)
	}
	formatted, err := Format(file)
	if err != nil {
		t.Fatal(err)
	}
	back, d := ParseFile("scalar.gooo", formatted)
	if d.HasErrors() {
		t.Fatal(d)
	}
	a := file.Declarations[2].(*ActivityDecl)
	b := back.Declarations[2].(*ActivityDecl)
	if a.Output != "Integer" || b.Output != "Integer" || !reflect.DeepEqual(a.Assembly.Spec, b.Assembly.Spec) {
		t.Fatal("source meaning changed", a, b)
	}
	clone := file.Clone().Declarations[2].(*ActivityDecl)
	clone.Assembly.Spec.Criteria.Inputs[0] = "[5]"
	if a.Assembly.Spec.Criteria.Inputs[0] != "[11]" {
		t.Fatal("source aliases criteria inputs")
	}
}
