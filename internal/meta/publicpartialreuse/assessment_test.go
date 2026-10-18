package publicpartialreuse

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestGoooPartialAssessmentAuthoredCases(t *testing.T) {
	raw, err := os.ReadFile("testdata/assessments.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string      `json:"name"`
		Input    Observation `json:"input"`
		Expected Assessment  `json:"expected"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil || len(cases) != 24 {
		t.Fatal("expected twenty-four independent authored assessments", err)
	}
	policy := receiptEvaluationFixture(t).Policy
	for _, item := range cases {
		t.Run(item.Name, func(t *testing.T) {
			value, actual, err := policy.Assess(context.Background(), item.Input)
			if err != nil || !reflect.DeepEqual(actual, item.Expected) {
				t.Fatalf("assessment differs: got %+v; want %+v; error %v", actual, item.Expected, err)
			}
			if value.ActivityID != AssessmentActivityID || value.SourceSHA256 != "sha256:"+policy.DecisionSourceDigest {
				t.Fatal("assessment lost its source", value)
			}
		})
	}
}

func TestGoooPartialAssessmentSourceOwnsDecision(t *testing.T) {
	input := receiptEvaluationFixture(t)
	delete(input.Receipts, "orders")
	observation, _, _ := observe(input)
	before, original, err := input.Policy.Assess(context.Background(), observation)
	if err != nil || original.Decision != DecisionUnknown {
		t.Fatal("expected unresolved missing receipt", original, err)
	}
	changed := bytes.Replace(input.Policy.decisionSource, []byte(`state == "INVALID"`), []byte(`state != "VALID"`), 1)
	if bytes.Equal(changed, input.Policy.decisionSource) {
		t.Fatal("source condition missing")
	}
	input.Policy.decisionSource = changed
	if _, _, err := input.Policy.Assess(context.Background(), observation); err == nil {
		t.Fatal("changed source retained an old identity")
	}
	input.Policy.DecisionSourceDigest = cache.HashBytes(changed).String()
	after, altered, err := input.Policy.Assess(context.Background(), observation)
	if err != nil || altered.Decision != DecisionRefuted || before.SourceSHA256 == after.SourceSHA256 ||
		before.ProjectionSHA256 == after.ProjectionSHA256 || before.ActivityID != after.ActivityID {
		t.Fatal("source-only condition did not change decision and origins", altered, err)
	}
}

func TestGoooPartialAssessmentFailureHasNoGoFallback(t *testing.T) {
	input := receiptEvaluationFixture(t)
	input.Policy.decisionSource = bytes.Replace(input.Policy.decisionSource, []byte("return Refuted()"), []byte("return Missing()"), 1)
	input.Policy.DecisionSourceDigest = cache.HashBytes(input.Policy.decisionSource).String()
	if report, err := Evaluate(input); err == nil || report.Decision != "" || report.PolicyEvaluation != nil {
		t.Fatal("invalid source became a normal decision", report, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	policy := receiptEvaluationFixture(t).Policy
	if value, _, err := policy.Assess(ctx, Observation{}); err == nil || value.ActivityID != "" {
		t.Fatal("canceled evaluation emitted a value", value, err)
	}
}

func TestGoooPartialAssessmentSavedReportProjection(t *testing.T) {
	for _, change := range []struct {
		name   string
		mutate func(*CaseReport)
	}{
		{"old-schema", func(r *CaseReport) { r.Schema = "gooo/public-partial-test-reuse-report/v1" }},
		{"missing-value", func(r *CaseReport) { r.PolicyEvaluation = nil }},
		{"source", func(r *CaseReport) { r.PolicyEvaluation.SourceSHA256 = "sha256:" + cache.HashBytes(nil).String() }},
		{"activity", func(r *CaseReport) { r.PolicyEvaluation.ActivityID += "-other" }},
		{"projection", func(r *CaseReport) { r.PolicyEvaluation.ProjectionSHA256 = "" }},
		{"input", func(r *CaseReport) { r.PolicyInputs.ReceiptStates[0] = "MISSING" }},
		{"value", func(r *CaseReport) { r.PolicyEvaluation.Value = []byte(`{}`) }},
		{"decision", func(r *CaseReport) { r.Decision = DecisionRefuted }},
		{"reason", func(r *CaseReport) { r.Reason = "invented" }},
		{"reused-count", func(r *CaseReport) { r.After.TestUnitsReused = 0 }},
	} {
		t.Run(change.name, func(t *testing.T) {
			report, err := Evaluate(receiptEvaluationFixture(t))
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(report)
			var saved CaseReport
			if err := json.Unmarshal(raw, &saved); err != nil || VerifyAssessment(saved) != nil {
				t.Fatal("unchanged saved assessment did not verify", err)
			}
			change.mutate(&saved)
			if err := VerifyAssessment(saved); err == nil {
				t.Fatal("changed report was accepted")
			}
		})
	}
}
