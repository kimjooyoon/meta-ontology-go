package sourceimpact

import (
	"context"
	"strings"
	"testing"
)

func TestCriteriaQuestionsImpactCountsSourceQueries(t *testing.T) {
	source := `package questions
namespace questions
entity Integer id "questions://integer"
entity Boolean id "questions://boolean"
entity Box id "questions://box" fields { field value id "questions://value" type integer required one }
activity Select(Integer) -> Box id "questions://select" computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[0]" holdout_input "[9]" }
 observe { criteria_input "[7]" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean id "questions://negative" computes "return input1.value == 0-input0"`
	changed := strings.Replace(source, `criteria_input "[7]"`, `criteria_input "[11]" criteria_input "[-5]"`, 1)
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(changed))
	if err != nil || len(r.Changes) != 1 || len(r.Affected) != 1 || r.ModelCalls != 0 || r.CandidateExecutions != 0 {
		t.Fatal(err, r)
	}
	row := r.Affected[0]
	if row.BeforeObligations.ObservationInputs != 1 || row.AfterObligations.ObservationInputs != 2 || row.AfterObligations.CriteriaInputs != 1 {
		t.Fatal("questions conflated with predicate-evaluated inputs", row)
	}
}
