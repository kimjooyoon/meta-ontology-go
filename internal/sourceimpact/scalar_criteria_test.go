package sourceimpact

import (
	"context"
	"strings"
	"testing"
)

func TestScalarCriteriaImpactRetainsPredicateAndQuestionCounts(t *testing.T) {
	source := `package feecriteria
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
	changed := strings.Replace(source, `criteria_input "[17]"`, `criteria_input "[18]" criteria_input "[-7]"`, 1)
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(changed))
	if err != nil || len(r.Changes) != 1 || len(r.Affected) != 1 || r.ModelCalls != 0 || r.CandidateExecutions != 0 {
		t.Fatal(err, r)
	}
	row := r.Affected[0]
	if row.BeforeObligations.ObservationInputs != 1 || row.AfterObligations.ObservationInputs != 2 || row.AfterObligations.CriteriaInputs != 1 {
		t.Fatal(row)
	}
}
