package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func generateCriteria(t *testing.T, source string) Result {
	t.Helper()
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "criteria.gooo", []byte(source), "Select")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestCriteriaRuntimeSelectsWithoutInventingExpectedValue(t *testing.T) {
	r := generateCriteria(t, typedCriteriaSource)
	a := r.Report.RecordAssembly
	if a.SelectedMask != 1 || a.Status != "COMPLETE_FINITE" || a.Passed != 0 || a.Total != 0 ||
		a.FieldsTotal != 0 || a.Criteria.Passed != 1 || a.Criteria.Total != 1 || a.Criteria.Unknown != 0 ||
		a.Criteria.Inputs[0].ActualJSON != `{"value":-7}` || a.CriteriaHoldout.Passed != 1 ||
		a.CriteriaHoldout.Inputs[0].ActualJSON != `{"value":-9}` || a.ModelCalls != 0 || len(a.Attempts) != 2 {
		t.Fatalf("criteria-only evidence differs: %+v", a)
	}
	if a.Attempts[0].Criteria.Failed != 1 || a.Attempts[0].Criteria.Inputs[0].Checks[0].Status != "UNSATISFIED" {
		t.Fatal("false was not preserved")
	}
	encoded, _ := json.MarshalIndent(r, "", "  ")
	var decoded Result
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "criteria.gooo", []byte(typedCriteriaSource), decoded); err != nil {
		t.Fatal("criteria receipt did not replay after JSON roundtrip", err)
	}
	candidate, _, err := RealizeRecordCandidate(context.Background(), "criteria.gooo", []byte(typedCriteriaSource), "Select", 0)
	if err != nil || candidate.Attempt.Criteria.Failed != 1 || candidate.Attempt.Total != 0 {
		t.Fatal(candidate, err)
	}
}

func TestCriteriaRuntimePreservesPartialAndUnknownPredicates(t *testing.T) {
	source := strings.Replace(typedCriteriaSource, `predicate "negative" activity "Negative"`,
		`predicate "negative" activity "Negative" predicate "positive" activity "Positive" predicate "broken" activity "Broken"`, 1) + `
activity Positive(Integer,Box) -> Boolean computes "return input1.value > 0"
activity Broken(Integer,Box) -> Boolean computes "return input0 / (input0-input0) > 0"`
	r := generateCriteria(t, source).Report.RecordAssembly
	if r.Status != "PARTIAL_FINITE" || r.Criteria.Passed != 1 || r.Criteria.Failed != 1 || r.Criteria.Unknown != 1 ||
		r.Criteria.Total != 3 || len(r.Attempts) != 2 || r.SelectedMask != 0 {
		t.Fatalf("partial/unknown evidence differs: %+v", r)
	}
	if !strings.Contains(r.Criteria.Inputs[0].Checks[2].Reason, "PREDICATE_UNAVAILABLE") {
		t.Fatal("predicate fault became false")
	}
}

func TestCriteriaRuntimeCandidateFaultDoesNotStopRemainingMasks(t *testing.T) {
	source := strings.Replace(typedCriteriaSource, "value:input", "value:input/(input-input)", 1)
	r := generateCriteria(t, source).Report.RecordAssembly
	if r.SelectedMask != 1 || r.Criteria.Passed != 1 || len(r.Attempts) != 2 || r.Attempts[0].Criteria.Unknown != 1 ||
		r.Attempts[0].Criteria.InputsCompleted != 0 || r.Attempts[0].Criteria.Inputs[0].ActualJSON != "" {
		t.Fatalf("candidate fault lost: %+v", r)
	}
	bad := strings.Replace(source, `alternative "0-input"`, `alternative "input/(input-input)"`, 1)
	g, _ := NewTypedPathGenerator("")
	_, err := g.GenerateSourceAssembly(context.Background(), "criteria.gooo", []byte(bad), "Select")
	var rejected *RecordAssemblyError
	if !errors.As(err, &rejected) || len(rejected.Receipt.Attempts) != 2 || rejected.Receipt.Attempts[0].Criteria.Unknown != 1 {
		t.Fatal("all failing candidates lost their unknown observations", err)
	}
}

func TestCriteriaRuntimeDoesNotStopAtExactCaseSuccess(t *testing.T) {
	source := strings.Replace(typedCriteriaSource, `attempts "2"`, `value_case "[0]" -> "{\"value\":0}" attempts "2"`, 1)
	r := generateCriteria(t, source).Report.RecordAssembly
	if r.SelectedMask != 1 || len(r.Attempts) != 2 || r.Passed != 1 || r.Total != 1 || r.Criteria.Passed != 1 {
		t.Fatal("exact case success bypassed criteria", r)
	}
	// Held input intentionally disagrees with the selected result. It must not
	// change selection or cause another candidate to be attempted.
	held := strings.Replace(typedCriteriaSource, "return input1.value == 0-input0", "if input0 == 9 { return input1.value == input0 }; return input1.value == 0-input0", 1)
	r = generateCriteria(t, held).Report.RecordAssembly
	if r.SelectedMask != 1 || r.Criteria.Passed != 1 || r.CriteriaHoldout.Failed != 1 || r.Status != "COMPLETE_FINITE" {
		t.Fatal("held result changed selection", r)
	}
}

func TestCriteriaRuntimeRejectsTamperedEvidence(t *testing.T) {
	result := generateCriteria(t, typedCriteriaSource)
	for _, mutate := range []func(*RecordAssemblyReceipt){
		func(r *RecordAssemblyReceipt) { r.Criteria = nil },
		func(r *RecordAssemblyReceipt) { r.Criteria.Passed++ },
		func(r *RecordAssemblyReceipt) { r.Criteria.Inputs[0].ActualJSON = `{"value":0}` },
		func(r *RecordAssemblyReceipt) { r.Criteria.Bindings[0].GeneratedSHA256 = "changed" },
		func(r *RecordAssemblyReceipt) { r.Attempts[0].Criteria.Unknown++ },
		func(r *RecordAssemblyReceipt) { r.CriteriaHoldout = nil },
		func(r *RecordAssemblyReceipt) { r.CriteriaHoldout.Inputs[0].Checks[0].Status = "UNSATISFIED" },
	} {
		raw, _ := json.Marshal(result)
		var changed Result
		json.Unmarshal(raw, &changed)
		mutate(changed.Report.RecordAssembly)
		if _, err := RealizeSourceAssembly(context.Background(), "criteria.gooo", []byte(typedCriteriaSource), changed); err == nil {
			t.Fatal("modified criteria evidence replayed")
		}
	}
}
