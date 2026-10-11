package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestIRBodyFillUsesSourcePureHelpers(t *testing.T) {
	for _, tc := range []struct{ name, helper, body, expression string }{
		{"skeleton", `activity Step(Integer) -> Integer computes "return input + 1"`,
			"return Step(__GOOO_BODY_HOLE_value__)", "input"},
		{"candidate", `activity Step(Integer) -> Integer computes "return input + 1"`,
			"return __GOOO_BODY_HOLE_value__", "Step(input)"},
		{"record helper", `activity Step(Integer) -> Box computes "return Box{value: input + 1}"`,
			"let box = Step(__GOOO_BODY_HOLE_value__); return box.value", "input"},
		{"multiple arguments", `activity Step(Integer, Box) -> Integer computes "return input0 + input1.value"`,
			"return Step(__GOOO_BODY_HOLE_value__, Box{value: 1})", "input"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(fmt.Sprintf(`package fillcalls
namespace fillcalls
entity Integer id "fillcalls://integer"
entity Box id "fillcalls://box" fields { field value id "fillcalls://value" type integer required one }
%s
activity Build(Integer) -> Integer computes %q
`, tc.helper, tc.body))
			plan := tinyBodyFillPlan("Add one using the source helper.", "value",
				IRBodyFillCandidate{ID: "wrong", Expression: "0"},
				IRBodyFillCandidate{ID: "helper", Expression: tc.expression})
			plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 3}, {Input: 10, Expected: 11}}
			plan.HoldoutTestCases = []IRBodyFillTestCase{{Input: 100, Expected: 101}}
			ctx := context.Background()
			r, err := GenerateWithIRBodyFill(ctx, "calls.gooo", source, "Build", plan, "", "")
			if err != nil {
				t.Fatal("fill scoring omitted source helper", err)
			}
			fill := r.Report.BodyFill
			if fill.LocalCalls == nil || fill.LocalCalls.Status != "RECORDED" {
				t.Fatal("typed helper closure was not observed", fill.LocalCalls)
			}
			if fill.SelectedCandidateID != "helper" || fill.TestCasesPassed != 2 || fill.HoldoutCasesPassed != 1 ||
				r.Report.CallClosure == nil || len(r.Report.CallClosure.Activities) != 1 ||
				fill.LocalModelPredictions != 0 || fill.ExternalProviderCalls != 0 {
				t.Fatal("helper selection or final closure differs", fill, r.Report.CallClosure)
			}
			raw, _ := json.MarshalIndent(r, "", "  ")
			var saved Result
			if err := json.Unmarshal(raw, &saved); err != nil {
				t.Fatal(err)
			}
			if _, err := ReplayIRBodyFill(ctx, "calls.gooo", source, plan, saved); err != nil {
				t.Fatal("saved fill helper did not replay", err)
			}
			changed := []byte(strings.Replace(string(source), tc.helper, tc.helper+" ", 1))
			if _, err := ReplayIRBodyFill(ctx, "calls.gooo", changed, plan, saved); err == nil {
				t.Fatal("changed source accepted with old fill receipt")
			}
		})
	}
}

func TestSourceRecordFillUsesCandidatePureHelper(t *testing.T) {
	source := []byte(`package fillcalls
namespace fillcalls
entity Integer id "fillcalls://integer"
entity Box id "fillcalls://box" fields { field value id "fillcalls://value" type integer required one }
activity Step(Integer) -> Integer computes "return input + 1"
activity Build(Box) -> Box computes "return Box{value: __GOOO_BODY_HOLE_value__ + __GOOO_BODY_HOLE_bias__}" assembling {
 source_fill intent "Add one to the field." {
  hole "value"
  hole "bias"
  candidate "identity" { fill "value" "input.value" fill "bias" "0" }
  candidate "helper" { fill "value" "Step(input.value)" fill "bias" "0" }
 }
 value_case "[{\"value\":2}]" -> "{\"value\":3}"
 holdout_value_case "[{\"value\":100}]" -> "{\"value\":101}"
}
`)
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "record.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateSourceAssembly(ctx, "record.gooo", source, "Build"); err != nil {
		t.Fatal("source fill preflight omitted helper", err)
	}
	r, err := GenerateWithSourceIRBodyFill(ctx, "record.gooo", source, "Build", spec, "", "", IRBodyFillOptions{})
	if err != nil || r.Report.BodyFill == nil || r.Report.BodyFill.SelectedCandidateID != "helper" || r.Report.BodyFill.TestCasesPassed != 1 ||
		r.Report.BodyFill.HoldoutCasesPassed != 1 || r.Report.CallClosure == nil {
		t.Fatal("record candidate helper was not selected", r.Report.BodyFill, err)
	}
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, r); err != nil {
		t.Fatal("record helper receipt did not replay", err)
	}
}

func TestIRBodyFillKeepsPureCallRestrictions(t *testing.T) {
	for _, helper := range []string{"", `activity Step(Integer) -> Integer computes "return Step(input)"`} {
		source := []byte("package invalid\nnamespace invalid\nentity Integer id \"invalid://integer\"\n" +
			helper + "\nactivity Build(Integer) -> Integer computes \"return Step(__GOOO_BODY_HOLE_value__)\"\n")
		plan := tinyBodyFillPlan("Choose a bounded helper argument.", "value",
			IRBodyFillCandidate{ID: "input", Expression: "input"},
			IRBodyFillCandidate{ID: "zero", Expression: "0"})
		if r, err := GenerateWithIRBodyFill(context.Background(), "invalid.gooo", source, "Build", plan, "", ""); err == nil || r.Source != "" {
			t.Fatal("unknown or recursive helper escaped pure-call checks", err)
		}
	}
}
