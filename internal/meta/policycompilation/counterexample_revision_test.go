package policycompilation

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

func counterexampleRevisionFixture(t *testing.T) ([]byte, PolicyRevisionCounterexample, CompiledPolicy) {
	t.Helper()
	source, request := revisionObservationFixture(t)
	proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
		ExpectedSourceDigest: request.ExpectedSourceDigest, Condition: request.Condition,
		FromDecision: request.FromDecision, ToDecision: request.ToDecision,
	})
	if err != nil {
		t.Fatal(err)
	}
	input := request.Cases[0].Baseline
	input.ValidatorExpectation = DecisionFailClosed
	return source, PolicyRevisionCounterexample{
		ExpectedSourceDigest: DigestBytes(source), Input: input,
		Observed: EvaluateSourcePolicy(proposal.Original, input),
	}, proposal.Original
}

func TestCounterexampleRevisionDerivesRequestFromGeneratedResult(t *testing.T) {
	source, counterexample, original := counterexampleRevisionFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	results, err := ExecuteGeneratedBatch(ctx, GenerateJudge(original), []Case{counterexample.Input})
	if err != nil || len(results) != 1 {
		t.Fatalf("observe original generated result: %v", err)
	}
	counterexample.Observed = results[0]
	sourceBefore, inputBefore := string(source), counterexample
	report, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", counterexample)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", counterexample)
	if err != nil || !reflect.DeepEqual(report, replay) {
		t.Fatalf("proposal replay differs: %v", err)
	}
	if report.State != "PROPOSED" || report.Revision == nil || report.CandidatePolicy == nil ||
		report.Revision.Condition != results[0].MatchedCondition || report.Revision.FromDecision != results[0].Decision ||
		report.Revision.ToDecision != counterexample.Input.ValidatorExpectation || len(report.DerivedFields) != 3 ||
		len(report.ChangedCoordinates) != 1 || report.CandidateSource == "" {
		t.Fatalf("request was not derived from its exact counterexample: %+v", report)
	}
	if string(source) != sourceBefore || !reflect.DeepEqual(counterexample, inputBefore) ||
		!reflect.DeepEqual(report.Counterexample, inputBefore) || report.Admission.State != "UNKNOWN" ||
		report.Improvement != "UNKNOWN" || report.CandidateExecution != "NOT_OBSERVED" ||
		report.MutationAuthority != 0 || report.PromotionAuthority != 0 {
		t.Fatal("proposal rewrote evidence or acquired execution/adoption authority")
	}
	stale, err := ExecuteGeneratedBatch(ctx, GenerateJudge(*report.CandidatePolicy), []Case{counterexample.Input})
	if err != nil || len(stale) != 1 || stale[0].MatchedCondition != ConditionSourceMismatch || stale[0].Decision != DecisionFailClosed {
		t.Fatalf("original evidence was silently rebound to the candidate: %v %+v", err, stale)
	}
}

func TestCounterexampleRevisionPreservesMissingStaleAndInvalidBoundaries(t *testing.T) {
	source, original, _ := counterexampleRevisionFixture(t)
	stale := DigestBytes([]byte("different observation"))
	cases := []struct {
		name   string
		mutate func(*PolicyRevisionCounterexample)
		state  string
		class  string
	}{
		{"missing-pin", func(c *PolicyRevisionCounterexample) { c.ExpectedSourceDigest = "" }, "UNKNOWN", "DIRECT_MISSING"},
		{"wrong-pin", func(c *PolicyRevisionCounterexample) { c.ExpectedSourceDigest = stale }, "REFUTED", ""},
		{"missing-producer", func(c *PolicyRevisionCounterexample) { c.Input.ProducerAvailable = false }, "UNKNOWN", "DIRECT_MISSING"},
		{"missing-consumer", func(c *PolicyRevisionCounterexample) { c.Input.ConsumerAvailable = false }, "UNKNOWN", "DIRECT_MISSING"},
		{"missing-provenance", func(c *PolicyRevisionCounterexample) { c.Input.Provenance = "" }, "UNKNOWN", "DIRECT_MISSING"},
		{"missing-source", func(c *PolicyRevisionCounterexample) { c.Input.ObservedSourceDigest = "" }, "UNKNOWN", "DIRECT_MISSING"},
		{"missing-independent", func(c *PolicyRevisionCounterexample) { c.Input.ObservedIndependentDigest = "" }, "UNKNOWN", "DIRECT_MISSING"},
		{"stale-source", func(c *PolicyRevisionCounterexample) { c.Input.ObservedSourceDigest = stale }, "UNKNOWN", "STALE"},
		{"stale-judge", func(c *PolicyRevisionCounterexample) { c.Input.ObservedGeneratedJudgeDigest = stale }, "UNKNOWN", "STALE"},
		{"missing-result", func(c *PolicyRevisionCounterexample) { c.Observed.Decision = "" }, "UNKNOWN", "DIRECT_MISSING"},
		{"upstream-unknown", func(c *PolicyRevisionCounterexample) { c.Observed.Decision = DecisionUnknown }, "UNKNOWN", "DEPENDENCY_BLOCKED"},
		{"unknown-decision", func(c *PolicyRevisionCounterexample) { c.Observed.Decision = "FIXED_POINT" }, "REFUTED", ""},
		{"different-case", func(c *PolicyRevisionCounterexample) { c.Observed.CaseID = "unrelated" }, "REFUTED", ""},
		{"unknown-condition", func(c *PolicyRevisionCounterexample) { c.Observed.MatchedCondition = "UNREGISTERED" }, "REFUTED", ""},
		{"unknown-expectation", func(c *PolicyRevisionCounterexample) { c.Input.ValidatorExpectation = "SUCCESS" }, "REFUTED", ""},
	}
	for _, current := range cases {
		t.Run(current.name, func(t *testing.T) {
			input := original
			current.mutate(&input)
			report, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", input)
			if err == nil || report.State != current.state || report.Revision != nil || report.CandidateSource != "" {
				t.Fatalf("invalid evidence produced a proposal: %+v error=%v", report, err)
			}
			if current.state == "UNKNOWN" && (report.Pending == nil || report.Pending.Stage == "" ||
				report.Pending.Step == "" || report.Pending.Reason == "" || report.Pending.UnknownClass != current.class ||
				report.Pending.NextOperation == "" || report.Pending.BlockedBy == nil) {
				t.Fatalf("UNKNOWN lost its cause: %+v", report.Pending)
			}
			if current.class == "DEPENDENCY_BLOCKED" && len(report.Pending.BlockedBy) != 1 {
				t.Fatal("upstream UNKNOWN lost its dependency")
			}
		})
	}
}

func TestCounterexampleRevisionDoesNotCallAgreementAFixedPoint(t *testing.T) {
	source, input, _ := counterexampleRevisionFixture(t)
	input.Input.ValidatorExpectation = input.Observed.Decision
	report, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", input)
	if err != nil || report.State != "NOT_PROPOSED" || report.Reason != "NO_DECISION_DIFFERENCE_DECLARED" ||
		report.Revision != nil || report.CandidateSource != "" || report.Improvement != "UNKNOWN" {
		t.Fatalf("declared agreement acquired an improvement claim: %+v %v", report, err)
	}
}

func TestDecodePolicyRevisionCounterexampleIsStrict(t *testing.T) {
	_, input, _ := counterexampleRevisionFixture(t)
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodePolicyRevisionCounterexample(raw)
	if err != nil || !reflect.DeepEqual(decoded, input) {
		t.Fatalf("valid counterexample did not roundtrip: %v", err)
	}
	for name, bad := range map[string]string{
		"null": "null", "trailing": string(raw) + "{}",
		"unknown": strings.TrimSuffix(string(raw), "}") + ",\"adopt\":true}",
		"duplicate": strings.Replace(string(raw), "\"expected_source_digest\":", "\"expected_source_digest\":\"x\",\"expected_source_digest\":", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodePolicyRevisionCounterexample([]byte(bad)); err == nil {
				t.Fatal("ambiguous counterexample was accepted")
			}
		})
	}
}
