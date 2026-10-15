package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestScalarCriteriaCoversSourceControlAndVariableChoices(t *testing.T) {
	for _, tc := range []struct {
		name, body, choice, predicate string
		picked                        uint16
	}{
		{"operand", "return 11-input", `choice "x" operand_order at "0" intent "subtract 11"`, "return input1 == input0-11", 1},
		{"branch", "if input < 0 { return input+11 } else { return input-11 }", `choice "x" branch_layout at "0" intent "reverse branch"`, "return input1 == input0+11", 1},
		{"reference", "let first=input-11; let second=input+11; return first", `choice "x" local_reference at "0" alternative "second" intent "choose second"`, "return input1 == input0+11", 1},
		{"assignment", "let first=input-11; let second=input+11; first=0; return second", `choice "x" assignment_target at "0" alternative "second" intent "set result to zero"`, "return input1 == 0", 1},
		{"order", "let first=input; let second=input; first=first+11; first=first*2; return first", `choice "x" root_order at "2" intent "multiply before adding"`, "return input1 == input0*2+11", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fmt.Sprintf(`package scalar
namespace scalar
entity Integer id "scalar://integer"
entity Boolean id "scalar://boolean"
activity Fee(Integer)->Integer computes %q assembling {
%s
criteria { predicate "required" activity "Required" input "[17]" holdout_input "[19]" }
attempts "2"
}
activity Required(Integer,Integer)->Boolean computes %q
`, tc.body, tc.choice, tc.predicate)
			r := generateScalarCriteria(t, source)
			a := r.Report.RecordAssembly
			if a.SelectedMask != tc.picked || a.Criteria.Passed != 1 || a.CriteriaHoldout.Passed != 1 || a.ModelCalls != 0 {
				t.Fatal("source path obligation not selected", a)
			}
			if _, err := RealizeSourceAssembly(context.Background(), "scalar.gooo", []byte(source), r); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestScalarCriteriaKeepsExactCasesAndHeldValuesSeparate(t *testing.T) {
	source := strings.Replace(scalarCriteriaSource, ` attempts "2"`, ` case "12" -> "1"
 holdout_case "21" -> "10"
 attempts "2"`, 1)
	r := generateScalarCriteria(t, source)
	a := r.Report.RecordAssembly
	if a.Passed != 1 || a.Total != 1 || a.FieldsTotal != 0 || a.Criteria.Passed != 2 || a.Criteria.Total != 2 ||
		a.Holdout == nil || a.Holdout.Passed != 1 || a.Holdout.Total != 1 || a.Holdout.FieldsTotal != 0 ||
		a.Holdout.Schema != "gooo/scalar-criteria-holdout/v1" || a.CriteriaHoldout.Passed != 1 {
		t.Fatal("case/criteria/held units mixed", a)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "scalar.gooo", []byte(source), r); err != nil {
		t.Fatal(err)
	}
	held, err := EvaluateSourceRecordHoldout(context.Background(), "scalar.gooo", []byte(source), []byte(r.GoooSource), "Fee")
	if err != nil || held == nil || held.Passed != 1 {
		t.Fatal("final scalar holdout missing", held, err)
	}
	for _, d := range r.Report.CompletenessReceipt.Dimensions {
		if strings.Contains(d.ID, "record_field") || strings.Contains(d.ID, "record_case") {
			t.Fatal("scalar reported as record", d)
		}
	}
}

func TestScalarCriteriaExplainsInsufficientChoicesAndUnknownPredicates(t *testing.T) {
	source := strings.Replace(scalarCriteriaSource, "return 11 - input", "return (11-input)+0", 1)
	a := generateScalarCriteria(t, source).Report.RecordAssembly
	reason := ExplainCriteriaRejections(a)
	if a.Status != "PARTIAL_FINITE" || a.CriteriaQuestions.Status != "NO_SURVIVING_CANDIDATE" ||
		a.Criteria.Passed != 1 || a.Criteria.Total != 2 || a.CriteriaHoldout.Failed != 1 || reason.Explained != 2 {
		t.Fatal("shared wrong scalar answers were accepted", a, reason)
	}
	unknown := strings.Replace(scalarCriteriaSource, "return input1 == input0 - 11", "return input0 / (input0-input0) == input1", 1)
	b := generateScalarCriteria(t, unknown).Report.RecordAssembly
	if b.Criteria.Unknown != b.Criteria.Total || b.Criteria.Failed != 0 || b.Status != "PARTIAL_FINITE" ||
		ExplainCriteriaRejections(b).Rejected != 0 {
		t.Fatal("unknown scalar predicate became false", b)
	}
}
