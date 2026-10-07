package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestReplayExternalBodyFillKeepsMultiHoleAndRecordPlans(t *testing.T) {
	for _, test := range []struct {
		name, source string
		plan         IRBodyFillPlan
	}{
		{"multi", "package fill\nnamespace fill\nentity Integer id \"fill://integer\"\nactivity Build(Integer) -> Integer computes \"let base = __GOOO_BODY_HOLE_base__; return base + __GOOO_BODY_HOLE_step__\"\n", IRBodyFillPlan{
			Schema: bodyFillMultiPlanSchema, Intent: "Add one.", Holes: []IRBodyFillHole{{ID: "base"}, {ID: "step"}},
			Candidates: []IRBodyFillCandidate{{ID: "add", Fills: map[string]string{"base": "input", "step": "1"}}, {ID: "subtract", Fills: map[string]string{"base": "input", "step": "-1"}}},
			TestCases:  []IRBodyFillTestCase{{Input: 2, Expected: 3}, {Input: 9007199254740993, Expected: 9007199254740994}},
		}},
		{"record", "package fill\nnamespace fill\nentity Input id \"fill://input\" fields { field value id \"fill://input/value\" type integer required one }\nentity Output id \"fill://output\" fields { field result id \"fill://output/result\" type integer required one }\nactivity Build(Input) -> Output computes \"return Output{result: __GOOO_BODY_HOLE_result__}\"\n", IRBodyFillPlan{
			Schema: bodyFillRecordPlanSchema, Intent: "Add one to the input field.", Holes: []IRBodyFillHole{{ID: "result"}},
			Candidates: []IRBodyFillCandidate{{ID: "add", Fills: map[string]string{"result": "input.value + 1"}}, {ID: "identity", Fills: map[string]string{"result": "input.value"}}},
			ValueCases: []assemblyspec.ValueCase{{Inputs: `[{"value":9007199254740993}]`, Expected: `{"result":9007199254740994}`}},
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			prior, err := GenerateWithIRBodyFill(ctx, "fill.gooo", []byte(test.source), "Build", test.plan, "", "")
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(prior)
			if err != nil {
				t.Fatal(err)
			}
			d := json.NewDecoder(strings.NewReader(string(raw)))
			d.UseNumber()
			if err := d.Decode(&prior); err != nil {
				t.Fatal(err)
			}
			completed, err := ReplayIRBodyFill(ctx, "fill.gooo", []byte(test.source), test.plan, prior)
			if err != nil || completed != prior.GoooSource {
				t.Fatal("external plan did not reconstruct its selected body", err)
			}
			prior.Report.BodyFill.CandidateScores[0].TestCasesPassed++
			if _, err := ReplayIRBodyFill(ctx, "fill.gooo", []byte(test.source), test.plan, prior); err == nil {
				t.Fatal("changed selection score replayed")
			}
		})
	}
}
