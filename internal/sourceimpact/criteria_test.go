package sourceimpact

import (
	"context"
	"strings"
	"testing"
)

func TestCriteriaImpactLinksPredicateToItsAssembly(t *testing.T) {
	source := `package criteria
namespace criteria
entity Integer id "criteria://integer"
entity Boolean id "criteria://boolean"
entity Box id "criteria://box" fields { field value id "criteria://value" type integer required one }
activity Select(Integer) -> Box id "criteria://select" computes "return Box{value:input}" assembling {
 choice "value" field_value at "0" alternative "0-input" intent "Negate."
 criteria { predicate "negative" activity "Negative" input "[7]" holdout_input "[9]" }
 attempts "2"
}
activity Negative(Integer,Box) -> Boolean id "criteria://negative" computes "return input1.value == 0-input0"`
	changed := strings.Replace(source, "input1.value == 0-input0", "input1.value == input0", 1)
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(changed))
	if err != nil || len(r.Changes) != 1 || len(r.Affected) != 2 || r.ModelCalls != 0 || r.CandidateExecutions != 0 {
		t.Fatal(err, r)
	}
	row := r.Affected[1]
	if row.ID != "criteria://select" || row.AfterObligations.CriteriaPredicates != 1 ||
		row.AfterObligations.CriteriaInputs != 1 || row.AfterObligations.CriteriaHoldouts != 1 ||
		row.Causes[0].Via.Kind != "assembly_criterion" || row.Causes[0].Via.Surface != "criteria:negative" {
		t.Fatal(row)
	}
	missing := strings.Replace(source, `activity "Negative"`, `activity "Missing"`, 1)
	unresolved, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(missing))
	if err != nil || len(unresolved.Unknowns) == 0 {
		t.Fatal("undeclared criterion reference was omitted", err)
	}
}
