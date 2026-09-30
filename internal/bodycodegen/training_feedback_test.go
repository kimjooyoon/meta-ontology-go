package bodycodegen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
)

func TestIRBodySearchExternalTrainingFeedbackRejectsStaleBindingsBeforeProvider(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}, {Input: 4, Expected: 4}}
	validFeedback := externalFeedbackFor(source, training, "identity", []IRBodyFillCaseResult{
		{Input: -1, Expected: 0, Actual: -1, Passed: false},
	})

	tests := []struct {
		name     string
		mutate   func(*ExternalTrainingFeedback)
		wantText string
	}{
		{name: "stale source", mutate: func(feedback *ExternalTrainingFeedback) { feedback.SourceDigest = digest([]byte("other source")) }, wantText: "source_digest"},
		{name: "stale suite", mutate: func(feedback *ExternalTrainingFeedback) { feedback.TrainingSuiteSHA256 = digest([]byte("other suite")) }, wantText: "training_suite_sha256"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			feedback := *validFeedback
			test.mutate(&feedback)
			plan := irBodySearchPlan([]IRBodyFillCandidate{
				{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"},
			}, training, nil, 2)
			plan.ExternalTrainingFeedback = &feedback
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { calls.Add(1) }))
			defer server.Close()
			_, err := GenerateWithIRBodySearch(context.Background(), "search.gooo", source, "ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
			if err == nil || !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("stale external feedback error = %v, want message containing %q", err, test.wantText)
			}
			if calls.Load() != 0 {
				t.Fatalf("provider received %d calls for invalid feedback", calls.Load())
			}
		})
	}
}

func TestIRBodySearchExternalTrainingFeedbackValidatesCandidateAndTrainingBindings(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}, {Input: 7, Expected: 0}}
	holdout := []IRBodyFillTestCase{{Input: 99, Expected: 0}}
	candidates := []IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}}

	tests := []struct {
		name         string
		candidateID  string
		observations []IRBodyFillCaseResult
	}{
		{name: "undeclared candidate", candidateID: "external-only", observations: []IRBodyFillCaseResult{{Input: -1, Expected: 0, Actual: -1}}},
		{name: "wrong expected", candidateID: "identity", observations: []IRBodyFillCaseResult{{Input: -1, Expected: 88, Actual: -1}}},
		{name: "duplicate case", candidateID: "identity", observations: []IRBodyFillCaseResult{{Input: -1, Expected: 0, Actual: -1}, {Input: -1, Expected: 0, Actual: -2}}},
		{name: "passed disagrees with actual", candidateID: "identity", observations: []IRBodyFillCaseResult{{Input: -1, Expected: 0, Actual: -1, Passed: true}}},
		{name: "holdout is not accepted", candidateID: "identity", observations: []IRBodyFillCaseResult{{Input: 99, Expected: 0, Actual: 99}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := irBodySearchPlan(candidates, training, holdout, 2)
			plan.ExternalTrainingFeedback = externalFeedbackFor(source, training, test.candidateID, test.observations)
			if err := validateIRBodySearchPlan(plan); err == nil {
				t.Fatal("invalid external observations were accepted")
			}
		})
	}
}

func TestExternalTrainingFeedbackJSONRequiresExplicitNonnullObservationFields(t *testing.T) {
	valid := `{"source_digest":"sha256:a","training_suite_sha256":"sha256:b","candidate_id":"identity","observations":[{"input":-1,"expected":0,"actual":-1,"passed":false}]}`
	for _, raw := range []string{
		strings.Replace(valid, `,"passed":false`, "", 1),
		strings.Replace(valid, `"actual":-1`, `"actual":null`, 1),
		strings.Replace(valid, `"expected":0`, `"expected":null`, 1),
		strings.Replace(valid, `"input":-1`, `"input":null`, 1),
		strings.Replace(valid, `"passed":false`, `"passed":null`, 1),
		strings.Replace(valid, `"input":-1`, `"input":-1,"input":-2`, 1),
		strings.Replace(valid, `"input":-1`, `"input":-1,"Input":-2`, 1),
		strings.Replace(valid, `"source_digest":"sha256:a"`, `"source_digest":"sha256:a","SOURCE_DIGEST":null`, 1),
		strings.Replace(valid, `"passed":false`, `"passed":false,"Passed":null`, 1),
	} {
		var plan IRBodySearchPlan
		wire := `{"schema":"` + bodySearchPlanSchema + `","intent":"x","hole_id":"floor","candidates":[{"id":"identity","expression":"input"},{"id":"zero","expression":"0"}],"test_cases":[{"input":-1,"expected":0}],"max_attempts":2,"external_training_feedback":` + raw + `}`
		if err := json.Unmarshal([]byte(wire), &plan); err == nil {
			t.Fatalf("accepted malformed external feedback observation: %s", raw)
		}
	}
	base := `{"schema":"` + bodySearchPlanSchema + `","intent":"x","hole_id":"floor","candidates":[{"id":"identity","expression":"input"},{"id":"zero","expression":"0"}],"test_cases":[{"input":-1,"expected":0}],"max_attempts":2,"external_training_feedback":` + valid + `}`
	for _, wire := range []string{
		strings.Replace(base, `"max_attempts":2`, `"max_attempts":2,"provider_model":"english","Provider_Model":null`, 1),
		strings.Replace(base, `"external_training_feedback":`+valid, `"external_training_feedback":`+valid+`,"External_Training_Feedback":null`, 1),
	} {
		var plan IRBodySearchPlan
		if err := json.Unmarshal([]byte(wire), &plan); err == nil {
			t.Fatalf("accepted case-insensitive duplicate plan key: %s", wire)
		}
	}
	validPlan := []byte(base)
	invalidUTF8 := append([]byte(nil), validPlan...)
	position := bytes.Index(invalidUTF8, []byte(`"intent":"x"`))
	if position < 0 {
		t.Fatal("failed to find intent string in test plan")
	}
	invalidUTF8[position+len(`"intent":"`)] = 0xff
	var decoded IRBodySearchPlan
	if err := json.Unmarshal(invalidUTF8, &decoded); err == nil {
		t.Fatal("accepted invalid UTF-8 in an external feedback plan")
	}
}

func TestIRBodySearchExternalTrainingFeedbackPromptIsCompactAndReceiptKeepsBindings(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := make([]IRBodyFillTestCase, 10)
	observations := make([]IRBodyFillCaseResult, 10)
	for i := range training {
		input := int64(-i - 1)
		training[i] = IRBodyFillTestCase{Input: input, Expected: 0}
		observations[i] = IRBodyFillCaseResult{Input: input, Expected: 0, Actual: input, Passed: false}
	}
	feedback := externalFeedbackFor(source, training, "identity", observations)
	plan := irBodySearchPlan([]IRBodyFillCandidate{
		{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}, {ID: "negate", Expression: "-input"},
	}, training, nil, 2)
	plan.PromptProfile = "compact"
	plan.ExternalTrainingFeedback = feedback

	var requests []searchRequestSnapshot
	server := newIRBodySearchServer(t, func(call int, snapshot searchRequestSnapshot) string {
		requests = append(requests, snapshot)
		return "zero"
	})
	defer server.Close()
	result, err := GenerateWithIRBodySearch(context.Background(), "search.gooo", source, "ClampNegativeToZero", plan, server.URL+"/v1/systemone", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(requests) != 1 || result.Report.BodySearch == nil || len(result.Report.BodySearch.Attempts) != 1 ||
		result.Report.BodySearch.Attempts[0].CandidateID != "zero" {
		t.Fatalf("external feedback caused a preselection score or extra attempt: requests=%d receipt=%#v", len(requests), result.Report.BodySearch)
	}
	var prompt struct {
		TrainingSuiteSHA256 string `json:"training_suite_sha256"`
		External            struct {
			CandidateID string `json:"candidate_id"`
			FailedCases []struct {
				Input    int64 `json:"input"`
				Expected int64 `json:"expected"`
				Actual   int64 `json:"actual"`
				Passed   *bool `json:"passed"`
			} `json:"failed_cases"`
			FailedCasesTotal     int  `json:"failed_cases_total"`
			FailedCasesTruncated bool `json:"failed_cases_truncated"`
		} `json:"external_training_feedback"`
		PriorAttempts []json.RawMessage `json:"prior_attempts"`
	}
	if err := json.Unmarshal([]byte(requests[0].Raw), &prompt); err != nil {
		t.Fatal(err)
	}
	if prompt.TrainingSuiteSHA256 != "" || prompt.External.CandidateID != "identity" || prompt.External.FailedCasesTotal != 10 ||
		!prompt.External.FailedCasesTruncated || len(prompt.External.FailedCases) != 8 || prompt.PriorAttempts == nil || len(prompt.PriorAttempts) != 0 {
		t.Fatalf("external feedback prompt was not compact or exposed a suite digest: %#v", prompt)
	}
	for i, failure := range prompt.External.FailedCases {
		if failure.Input != int64(-i-1) || failure.Expected != 0 || failure.Actual != int64(-i-1) || failure.Passed != nil {
			t.Fatalf("failure %d is not a compact training triple: %#v", i, failure)
		}
	}
	feedbackHash := feedbackDigest(t, feedback)
	for _, secret := range []string{feedback.SourceDigest, feedback.TrainingSuiteSHA256, feedbackHash} {
		if strings.Contains(requests[0].Raw, secret) {
			t.Fatalf("external feedback hash %q leaked into model state", secret)
		}
	}
	provenance := result.Report.BodySearch.ExternalTrainingFeedback
	if provenance == nil || provenance.FeedbackSHA256 != feedbackHash || provenance.SourceDigest != digest(source) ||
		provenance.TrainingSuiteSHA256 != digest(mustMarshal(t, training)) || provenance.CandidateID != "identity" ||
		provenance.ObservationCount != 10 || provenance.FailedObservationCount != 10 || provenance.PromptedFailureCount != 8 ||
		!provenance.FailedObservationsTruncated || result.Report.BodySearch.PromptProfile != "compact" {
		t.Fatalf("local receipt omitted external feedback provenance: %#v", provenance)
	}
	plan.ProviderModel = "typed-decisions"
	request, err := bodySearchRequest("ClampNegativeToZero", "stable-id", "body", plan,
		plan.Candidates, &IRBodySearchReceipt{TrainingSuiteSHA256: digest(mustMarshal(t, training))})
	if err != nil || request.ProviderModel != "typed-decisions" || !strings.Contains(request.Question.Instructions, "advisory observations") {
		t.Fatalf("provider model was not propagated to the decision request: request=%#v err=%v", request, err)
	}
}

func TestIRBodySearchCompactPromptProfileIsIndependentOfFeedbackPresence(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}}
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}}, training, nil, 2)
	plan.PromptProfile = "compact"
	receipt := &IRBodySearchReceipt{TrainingSuiteSHA256: digest(mustMarshal(t, training))}

	withoutFeedback, err := bodySearchRequest("ClampNegativeToZero", "activity-id", "body-ir", plan, plan.Candidates, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(withoutFeedback.State, "training_suite_sha256") || strings.Contains(withoutFeedback.State, "external_training_feedback") {
		t.Fatalf("compact no-feedback state retained a suite hash or hint field: %s", withoutFeedback.State)
	}
	plan.ExternalTrainingFeedback = externalFeedbackFor(source, training, "identity", []IRBodyFillCaseResult{
		{Input: -1, Expected: 0, Actual: -1, Passed: false},
	})
	withFeedback, err := bodySearchRequest("ClampNegativeToZero", "activity-id", "body-ir", plan, plan.Candidates, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(withFeedback.State, "training_suite_sha256") || !strings.Contains(withFeedback.State, "external_training_feedback") ||
		withoutFeedback.Question.Instructions != withFeedback.Question.Instructions {
		t.Fatalf("compact feedback arm changed packaging or instructions: without=%#v with=%#v", withoutFeedback, withFeedback)
	}
	for _, secret := range []string{plan.ExternalTrainingFeedback.SourceDigest, plan.ExternalTrainingFeedback.TrainingSuiteSHA256,
		feedbackDigest(t, plan.ExternalTrainingFeedback)} {
		if strings.Contains(withFeedback.State, secret) {
			t.Fatalf("compact feedback hash %q leaked into state", secret)
		}
	}
}

func TestIRBodySearchLegacyPromptProfileKeepsSuiteDigestWithExternalFeedback(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}}
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}}, training, nil, 2)
	plan.ExternalTrainingFeedback = externalFeedbackFor(source, training, "identity", []IRBodyFillCaseResult{
		{Input: -1, Expected: 0, Actual: -1, Passed: false},
	})
	receipt := &IRBodySearchReceipt{TrainingSuiteSHA256: digest(mustMarshal(t, training))}
	request, err := bodySearchRequest("ClampNegativeToZero", "activity-id", "body-ir", plan, plan.Candidates, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(request.State, receipt.TrainingSuiteSHA256) ||
		request.Question.Instructions != "Choose one untried expression to fill the Gooo body hole. Use the declared intent and previous training failures. This choice will be typechecked and tested after selection." {
		t.Fatalf("legacy feedback changed the prior prompt package: %#v", request)
	}
}

func TestIRBodySearchExternalFeedbackWithoutProviderRemainsDeclaredOrderDeterministic(t *testing.T) {
	source := readIRBodySearchFixture(t)
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}, {Input: 1, Expected: 1}}
	feedback := externalFeedbackFor(source, training, "zero", []IRBodyFillCaseResult{{Input: -1, Expected: 0, Actual: 0, Passed: true}})
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}}, training, nil, 1)
	plan.ExternalTrainingFeedback = feedback
	first, err := GenerateWithIRBodySearch(context.Background(), "search.gooo", source, "ClampNegativeToZero", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateWithIRBodySearch(context.Background(), "search.gooo", source, "ClampNegativeToZero", plan, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if first.Source != second.Source || first.Report.GeneratedDigest != second.Report.GeneratedDigest ||
		first.Report.BodySearch.Attempts[0].CandidateID != "identity" || second.Report.BodySearch.Attempts[0].CandidateID != "identity" ||
		first.Report.BodySearch.ProviderOperations != 0 || second.Report.BodySearch.ProviderOperations != 0 {
		t.Fatalf("external feedback changed no-provider declaration-order replay: first=%#v second=%#v", first.Report.BodySearch, second.Report.BodySearch)
	}
}

func TestIRBodySearchRejectsUnsupportedProviderModelBeforeProvider(t *testing.T) {
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "zero", Expression: "0"}, {ID: "identity", Expression: "input"}},
		[]IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)
	plan.ProviderModel = "model-that-is-not-supported"
	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("unsupported provider model was accepted")
	}
}

func TestIRBodySearchRejectsUnsupportedPromptProfile(t *testing.T) {
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "zero", Expression: "0"}, {ID: "identity", Expression: "input"}},
		[]IRBodyFillTestCase{{Input: -1, Expected: 0}}, nil, 2)
	plan.PromptProfile = "verbose-but-not-supported"
	if err := validateIRBodySearchPlan(plan); err == nil {
		t.Fatal("unsupported prompt profile was accepted")
	}
}

func TestIRBodySearchRequestBytesAndDigestStayStableWithoutNewPlanFields(t *testing.T) {
	training := []IRBodyFillTestCase{{Input: -1, Expected: 0}}
	plan := irBodySearchPlan([]IRBodyFillCandidate{{ID: "identity", Expression: "input"}, {ID: "zero", Expression: "0"}}, training, nil, 2)
	trainingBytes, _ := json.Marshal(training)
	receipt := &IRBodySearchReceipt{TrainingSuiteSHA256: digest(trainingBytes)}
	request, err := bodySearchRequest("ClampNegativeToZero", "activity-id", "body-ir", plan, plan.Candidates, receipt)
	if err != nil {
		t.Fatal(err)
	}
	legacyState, err := json.Marshal(struct {
		Schema            string                `json:"schema"`
		Stage             string                `json:"stage"`
		Activity          string                `json:"activity"`
		ActivityID        string                `json:"activity_id"`
		BodyIR            string                `json:"body_ir"`
		Intent            string                `json:"intent"`
		TrainingTestCount int                   `json:"training_test_count"`
		TrainingSuiteSHA  string                `json:"training_suite_sha256"`
		Remaining         []IRBodyFillCandidate `json:"remaining_candidates"`
		PriorAttempts     []bodySearchFeedback  `json:"prior_attempts"`
	}{"gooo/body-codegen-ir-search-state/v1", "choose_before_candidate_evaluation", "ClampNegativeToZero", "activity-id",
		"body-ir", plan.Intent, len(training), receipt.TrainingSuiteSHA256, plan.Candidates, []bodySearchFeedback{}})
	if err != nil {
		t.Fatal(err)
	}
	if request.State != string(legacyState) || request.ProviderModel != "" {
		t.Fatalf("request state changed when both new plan fields are absent: got=%s want=%s", request.State, legacyState)
	}
	legacyRequest, err := json.Marshal(struct {
		Schema   string                 `json:"schema"`
		State    string                 `json:"state"`
		Question decisionroute.Question `json:"question"`
		Fallback string                 `json:"fallback"`
	}{request.Schema, request.State, request.Question, request.Fallback})
	if err != nil {
		t.Fatal(err)
	}
	currentRequest, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if string(currentRequest) != string(legacyRequest) {
		t.Fatalf("typed decision request bytes changed without the new plan fields: got=%s want=%s", currentRequest, legacyRequest)
	}
	gotDigest, err := decisionroute.Validate(request)
	if err != nil || gotDigest != digest(legacyRequest) {
		t.Fatalf("decision digest changed without the new plan fields: got=%q err=%v want=%q", gotDigest, err, digest(legacyRequest))
	}
}

func externalFeedbackFor(source []byte, training []IRBodyFillTestCase, candidate string, observations []IRBodyFillCaseResult) *ExternalTrainingFeedback {
	trainingBytes, _ := json.Marshal(training)
	return &ExternalTrainingFeedback{
		SourceDigest: digest(source), TrainingSuiteSHA256: digest(trainingBytes),
		CandidateID: candidate, Observations: observations,
	}
}

func feedbackDigest(t *testing.T, feedback *ExternalTrainingFeedback) string {
	t.Helper()
	return digest(mustMarshal(t, feedback))
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(fmt.Errorf("marshal test value: %w", err))
	}
	return data
}
