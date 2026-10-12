package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestCriteriaPolicyAndResumePreservePredicateObligations(t *testing.T) {
	ctx := context.Background()
	g, _ := NewTypedPathGenerator("")
	source := []byte(typedCriteriaSource)
	stop := fixedPolicy(t, "PROGRESS", "USE_OBSERVED_CANDIDATE")
	prior, err := g.GenerateRecordAssemblyWithPolicy(ctx, "criteria.gooo", source, "Select", stop)
	if err != nil {
		t.Fatal(err)
	}
	old := prior.Report.RecordAssembly
	if old.Status != "PARTIAL_FINITE" || len(old.Attempts) != 1 || old.Criteria.Failed != 1 ||
		old.Control.Decisions[0].Input.Total != 1 || old.Control.Decisions[0].Input.Matched != 0 ||
		old.Control.Decisions[0].CountUnit == "" {
		t.Fatal("policy did not observe criteria input rows", old)
	}
	before, _ := json.Marshal(prior)
	result, err := ResumeRecordAssembly(ctx, "criteria.gooo", source, source, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Status != "COMPLETE_FINITE" || r.SelectedMask != 1 || r.Criteria.Passed != 1 ||
		r.Continuation.AddedAttempts != 1 || r.Continuation.NewModelCalls != 0 || r.Total != 0 {
		t.Fatal(r)
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("resume changed historical criteria observations")
	}
	if _, err := RealizeSourceAssembly(ctx, "criteria.gooo", source, result); err != nil {
		t.Fatal(err)
	}
}

func TestCriteriaDependencyChangeRechecksCompletedPrefix(t *testing.T) {
	ctx := context.Background()
	source := []byte(typedCriteriaSource)
	prior := generateCriteria(t, typedCriteriaSource)
	changed := []byte(strings.Replace(typedCriteriaSource, "return input1.value == 0-input0", "return input1.value == input0", 1))
	result, err := ResumeRecordAssembly(ctx, "criteria.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 0 || r.Criteria.Passed != 1 || r.Continuation.RecheckedAttempts != 2 || r.Continuation.AddedAttempts != 0 ||
		r.Criteria.Bindings[0].GeneratedSHA256 == prior.Report.RecordAssembly.Criteria.Bindings[0].GeneratedSHA256 ||
		len(r.ControlHistory) != 1 || r.ControlHistory[0].Source != string(source) {
		t.Fatal("changed criterion did not recheck original candidates", r)
	}
	if _, err := RealizeSourceAssembly(ctx, "criteria.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
}
