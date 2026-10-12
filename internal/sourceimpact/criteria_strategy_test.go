package sourceimpact

import (
	"context"
	"strings"
	"testing"
)

func TestCriteriaStrategyImpactTracksQuestionMeaning(t *testing.T) {
	source := `package challenge
namespace challenge
entity Integer id "challenge://integer"
entity Boolean id "challenge://boolean"
entity Box id "challenge://box" fields { field value id "challenge://value" type integer required one }
activity Choose(Integer) -> Box id "challenge://choose" computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "input+0" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[0]" }
 observe { criteria_input "[7]" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"`
	changed := strings.Replace(source, "observe {", `observe { criteria_strategy "distinguish_then_check"`, 1)
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(changed))
	if err != nil || len(r.Changes) != 1 || len(r.Affected) != 1 || r.Affected[0].ID != "challenge://choose" ||
		r.CandidateExecutions != 0 || r.ModelCalls != 0 {
		t.Fatal(err, r)
	}
	if r.Affected[0].BeforeObligations.ObservationInputs != 1 || r.Affected[0].AfterObligations.ObservationInputs != 1 {
		t.Fatal("strategy change invented input obligations", r.Affected)
	}
}
