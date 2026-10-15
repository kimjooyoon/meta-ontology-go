package sourceimpact

import (
	"context"
	"strings"
	"testing"
)

func TestRecordObservationImpactPreservesQuestionsAndOracle(t *testing.T) {
	source := `package recordimpact
namespace recordimpact
entity Integer id "recordimpact://integer"
entity Box id "recordimpact://box" fields { field value id "recordimpact://value" type integer required one }
activity Select(Integer) -> Box id "recordimpact://select" computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "0 - input" intent "Negate."
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[17]" value_input "[19]" oracle "Expected" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Expected(Integer) -> Box id "recordimpact://expected" computes "return Box{value: 0 - input}"`
	changed := strings.Replace(source, "return Box{value: 0 - input}", "return Box{value: input}", 1)
	r, err := Compare(context.Background(), "before.gooo", []byte(source), "after.gooo", []byte(changed))
	if err != nil || len(r.Changes) != 1 || len(r.Affected) != 2 || r.ModelCalls != 0 || r.CandidateExecutions != 0 {
		t.Fatal(err, r)
	}
	row := r.Affected[1]
	if row.ID != "recordimpact://select" || row.BeforeObligations.ObservationInputs != 2 || row.AfterObligations.ObservationInputs != 2 || row.AfterObligations.Oracle != "Expected" || row.Causes[0].Via.Kind != "observation_oracle" {
		t.Fatal(row)
	}
}
