package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const scalarCriteriaSource = `package feecriteria
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

func TestScalarCriteriaConstructionObservationExposesPredicateRows(t *testing.T) {
	suite, err := DecodeCompositionInputs([]byte(`{"schema":"gooo/body-composition-inputs/v1","inputs":[{"Fee":17}]}`))
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateCompositionWithOptions(context.Background(), "scalar.gooo", []byte(scalarCriteriaSource), suite, CompositionOptions{EntryActivity: "Fee"})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ObserveConstruction(context.Background(), []byte(scalarCriteriaSource), prior)
	if err != nil || len(rows) != 2 {
		t.Fatal(err, rows)
	}
	if rows[0].Counts.Matched != 1 || rows[0].Counts.Total != 2 || !rows[0].ScoringCompleted ||
		rows[1].Counts.Matched != 2 || rows[1].Counts.Total != 2 || rows[1].Profile != "scalar_criteria_choices" {
		t.Fatal(rows)
	}
}
func TestScalarCriteriaCallerCannotDiscardAcquiredFailure(t *testing.T) {
	source := strings.Replace(scalarCriteriaSource, "return 11 - input", "return (11-input)+0", 1) + `
activity Main(Integer)->Integer computes "return Fee(input+3)"`
	construction := jointCases(t, `[{"inputs":{"Main":8},"expected":{"Main":0}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":24},"expected":{"Main":-16}}]`)
	r, observed, err := ConstructAndEvaluateJoint(context.Background(), "scalar.gooo", []byte(source), construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || len(r.Attempts) != 2 || observed.Runtime.FinitePassed != 1 || observed.NewModelCalls != 0 {
		t.Fatal(r, observed)
	}
	for _, a := range r.Attempts {
		if a.Runtime.FinitePassed != 1 || a.CriteriaPassed != 1 || a.CriteriaFailed != 1 || a.CriteriaTotal != 2 {
			t.Fatal("caller match erased acquired scalar input", a)
		}
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(context.Background(), "scalar.gooo", []byte(source), saved, evaluation, nativeTool())
	if err != nil || !replay.ConstructionReplayed || replay.Runtime.FinitePassed != 1 || replay.NewModelCalls != 0 {
		t.Fatal(err, replay)
	}
}
