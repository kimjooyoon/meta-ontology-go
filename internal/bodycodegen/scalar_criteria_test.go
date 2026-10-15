package bodycodegen

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

func generateScalarCriteria(t *testing.T, source string) Result {
	t.Helper()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "scalar.gooo", []byte(source), "Fee")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestScalarCriteriaPreservesSignatureAndAcquiredObligations(t *testing.T) {
	r := generateScalarCriteria(t, scalarCriteriaSource)
	a := r.Report.RecordAssembly
	if r.Report.InputType != "int64" || r.Report.OutputType != "int64" || len(r.Report.RecordTypes) != 0 {
		t.Fatal("scalar signature changed", r.Report.InputType, r.Report.OutputType, r.Report.RecordTypes)
	}
	if a.Schema != scalarCriteriaAssemblySchema || a.SelectedMask != 1 || a.Status != "COMPLETE_FINITE" ||
		a.Criteria.Passed != 2 || a.Criteria.Total != 2 || a.Total != 0 || a.FieldsTotal != 0 ||
		a.CriteriaHoldout.Passed != 1 || a.CriteriaHoldout.Inputs[0].ActualJSON != "9" || a.ModelCalls != 0 {
		t.Fatalf("scalar obligations not carried: %+v", a)
	}
	if len(a.CriteriaQuestions.AcquiredInputs) != 1 || a.CriteriaQuestions.AcquiredInputs[0] != "[17]" ||
		!strings.Contains(r.GoooSource, `picked "direction" -> "layout_reverse"`) {
		t.Fatal("source acquisition/checkpoint missing", a.CriteriaQuestions, r.GoooSource)
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var restored Result
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "scalar.gooo", []byte(scalarCriteriaSource), restored); err != nil {
		t.Fatal("scalar receipt replay", err)
	}
	checkpoint := generateScalarCriteria(t, r.GoooSource)
	if checkpoint.Source != r.Source {
		t.Fatal("selected scalar checkpoint changed behavior")
	}
}
