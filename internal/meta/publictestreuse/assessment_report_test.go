package publictestreuse

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestRetainedGoooAssessmentRequiresExactRecords(t *testing.T) {
	if err := VerifyAssessment(retainedAssessmentFixture()); err != nil {
		t.Fatal("complete retained record was rejected", err)
	}
	cases := []struct {
		name   string
		change func(*Report)
	}{
		{"missing-evaluation", func(r *Report) { r.PolicyEvaluation = nil }},
		{"missing-evaluation-count", func(r *Report) { r.PolicyEvaluations = 0 }},
		{"extra-evaluation-count", func(r *Report) { r.PolicyEvaluations = 2 }},
		{"wrong-source", func(r *Report) { r.PolicyEvaluation.SourceSHA256 = cache.HashBytes([]byte("other")).String() }},
		{"missing-projection", func(r *Report) { r.PolicyEvaluation.ProjectionSHA256 = "" }},
		{"wrong-activity", func(r *Report) { r.PolicyEvaluation.ActivityID = "gooo://test-reuse/other" }},
		{"wrong-operation", func(r *Report) { r.Operation = OriginalOperation }},
		{"wrong-case", func(r *Report) { r.CaseID = CaseStaleEvidence }},
		{"wrong-decision", func(r *Report) { r.Decision = DecisionUnknown }},
		{"wrong-reason", func(r *Report) { r.Reason = ReasonStale }},
		{"wrong-reuse-count", func(r *Report) { r.ReusedTestExecutions = 0 }},
		{"extra-reuse-count", func(r *Report) { r.ReusedTestExecutions = 2 }},
		{"unknown-on-closed", func(r *Report) { r.Unknown = Unknown("x", "x", "x", "x") }},
		{"missing-input", func(r *Report) { r.PolicyInputs = json.RawMessage(`{}`) }},
		{"null-input", func(r *Report) { r.PolicyInputs = json.RawMessage(`null`) }},
		{"input-trailer", func(r *Report) { r.PolicyInputs = append(r.PolicyInputs, []byte(` {}`)...) }},
		{"input-alias", func(r *Report) { replaceAssessmentInput(r, `"requested"`, `"Requested"`) }},
		{"input-null-field", func(r *Report) { replaceAssessmentInput(r, `"requested":true`, `"requested":null`) }},
		{"input-wrong-type", func(r *Report) { replaceAssessmentInput(r, `"requested":true`, `"requested":"true"`) }},
		{"unrequested-reuse", func(r *Report) { replaceAssessmentInput(r, `"requested":true`, `"requested":false`) }},
		{"input-duplicate", func(r *Report) { replaceAssessmentInput(r, `"requested":true`, `"requested":false,"requested":true`) }},
		{"input-extra-field", func(r *Report) { replaceAssessmentInput(r, `"requested":true`, `"requested":true,"extra":1`) }},
		{"missing-result", func(r *Report) { r.PolicyEvaluation.Value = json.RawMessage(`{}`) }},
		{"null-result", func(r *Report) { r.PolicyEvaluation.Value = json.RawMessage(`null`) }},
		{"result-null-field", func(r *Report) { replaceAssessmentResult(r, `"reuse":true`, `"reuse":null`) }},
		{"result-alias", func(r *Report) { replaceAssessmentResult(r, `"reuse"`, `"Reuse"`) }},
		{"result-duplicate", func(r *Report) { replaceAssessmentResult(r, `"reuse":true`, `"reuse":false,"reuse":true`) }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			report := retainedAssessmentFixture()
			item.change(&report)
			if err := VerifyAssessment(report); err == nil {
				t.Fatal("inconsistent retained record was accepted")
			}
		})
	}
}

func TestRetainedOriginalAssessmentRequiresObservedSuccess(t *testing.T) {
	report := retainedAssessmentFixture()
	report.Operation, report.CaseID = OriginalOperation, CaseBaselineExecution
	report.Reason, report.ReusedTestExecutions = ReasonBaseline, 0
	report.PolicyEvaluation.ActivityID = "gooo://test-reuse/assess-original/v1"
	report.PolicyInputs = json.RawMessage(`{"build_passed":true,"test_passed":true,"semantic_matches":true}`)
	report.PolicyEvaluation.Value = json.RawMessage(`{"case_id":"BASELINE_TEST_EXECUTION","decision":"CLOSED","reason":"SUCCESSFUL_ORIGINAL_TEST_EXECUTION","step":"","next_operation":"","blocked_by":"","reuse":false}`)
	if err := VerifyAssessment(report); err != nil {
		t.Fatal("consistent original report was rejected", err)
	}
	replaceAssessmentInput(&report, `"build_passed":true`, `"build_passed":false`)
	if err := VerifyAssessment(report); err == nil {
		t.Fatal("unsuccessful observation was reported as completed")
	}
}

func TestRetainedGoooAssessmentPreservesNextOperation(t *testing.T) {
	report := retainedAssessmentFixture()
	report.CaseID, report.Decision, report.Reason = CaseMissingAuthorization, DecisionUnknown, ReasonMissingAuthorization
	report.ReusedTestExecutions = 0
	report.PolicyInputs = json.RawMessage(`{"requested":false,"receipt_state":"","policy_state":"","binding_state":"","input_profile":"gooo/generated-test-package-inputs/v1"}`)
	report.PolicyEvaluation.Value = json.RawMessage(`{"case_id":"MISSING_EXPLICIT_AUTHORIZATION","decision":"UNKNOWN","reason":"MISSING_EXPLICIT_REUSE_AUTHORIZATION","step":"AUTHORIZE_REUSE","next_operation":"PROVIDE_EXPLICIT_REUSE_AUTHORIZATION","blocked_by":"explicit_reuse_authorization","reuse":false}`)
	report.Unknown = Unknown(ReasonMissingAuthorization, UnknownAuthorization, UnknownNextAuthorization, "explicit_reuse_authorization")
	if err := VerifyAssessment(report); err != nil {
		t.Fatal("complete unresolved result was rejected", err)
	}
	report.Unknown.NextOperation = "RUN_TEST_WITHOUT_REQUEST"
	if err := VerifyAssessment(report); err == nil {
		t.Fatal("changed next operation was accepted")
	}
}

func retainedAssessmentFixture() Report {
	source := cache.HashBytes([]byte("retained fixture source")).String()
	return Report{
		Schema:                     ReportSchema,
		Operation:                  ReplayOperation,
		CaseID:                     CaseAuthorizedReuse,
		Decision:                   DecisionClosed,
		Reason:                     ReasonReuse,
		Binding:                    Binding{PolicySourceDigest: source},
		ReusedTestExecutions:       1,
		PolicyDecisionSourceDigest: source,
		PolicyEvaluations:          1,
		PolicyInputs:               json.RawMessage(`{"requested":true,"receipt_state":"VALID","policy_state":"MATCH","binding_state":"MATCH","input_profile":"gooo/generated-test-package-inputs/v1"}`),
		PolicyEvaluation:           &bodycodegen.FixedValue{
			ActivityID:       "gooo://test-reuse/assess-replay/v1",
			SourceSHA256:     source,
			ProjectionSHA256: cache.HashBytes([]byte("retained fixture projection")).String(),
			Value:            json.RawMessage(`{"case_id":"AUTHORIZED_IMMUTABLE_RECEIPT_REUSE","decision":"CLOSED","reason":"AUTHORIZED_IMMUTABLE_RECEIPT_REUSE","step":"","next_operation":"","blocked_by":"","reuse":true}`),
		},
	}
}

func replaceAssessmentInput(report *Report, before, after string) {
	report.PolicyInputs = bytes.Replace(report.PolicyInputs, []byte(before), []byte(after), 1)
}

func replaceAssessmentResult(report *Report, before, after string) {
	report.PolicyEvaluation.Value = bytes.Replace(report.PolicyEvaluation.Value, []byte(before), []byte(after), 1)
}
