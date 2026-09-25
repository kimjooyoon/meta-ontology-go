package policycompilation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func counterexampleExecutionFixture(t *testing.T) ([]byte, PolicyCounterexampleExecutionInput) {
	t.Helper()
	source, counterexample, _ := counterexampleRevisionFixture(t)
	proposal, err := ProposePolicyRevisionFromCounterexample("policy.gooo", source,
		"metapolicycompilation", "metapolicycompilation", counterexample)
	if err != nil {
		t.Fatal(err)
	}
	policy := proposal.CandidatePolicy
	fresh := Case{
		ID: counterexample.Input.ID, ValidatorExpectation: counterexample.Input.ValidatorExpectation,
		EvidenceClass: EvidenceSyntheticFixture, Provenance: "synthetic candidate snapshot, not independent evidence",
		ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
		ObservedGeneratedJudgeDigest: DigestBytes(GenerateJudge(*policy)), ObservedIndependentDigest: policy.SemanticDigest,
	}
	stale := counterexample.Input
	stale.ID = "stale"
	missing := Case{ID: "missing", EvidenceClass: EvidenceSyntheticFixture, Provenance: "explicit absent evidence"}
	return source, PolicyCounterexampleExecutionInput{Counterexample: counterexample, Cases: []PolicyRevisionCasePair{
		{Baseline: counterexample.Input, Candidate: fresh}, {Baseline: stale, Candidate: stale},
		{Baseline: missing, Candidate: missing},
	}}
}

func counterexampleExecutionJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestCounterexampleExecutionRejectsUnboundInputsBeforeNativeWork(t *testing.T) {
	source, fixture := counterexampleExecutionFixture(t)
	cases := map[string]func(*PolicyCounterexampleExecutionInput){
		"missing-pin": func(v *PolicyCounterexampleExecutionInput) { v.Counterexample.ExpectedSourceDigest = "" },
		"stale-evidence": func(v *PolicyCounterexampleExecutionInput) {
			v.Counterexample.Input.ObservedGeneratedJudgeDigest = DigestBytes([]byte("stale"))
		},
		"changed-expectation":    func(v *PolicyCounterexampleExecutionInput) { v.Cases[0].Candidate.ValidatorExpectation = DecisionPass },
		"missing-counterexample": func(v *PolicyCounterexampleExecutionInput) { v.Cases = v.Cases[1:] },
		"missing-pairs":          func(v *PolicyCounterexampleExecutionInput) { v.Cases = nil },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			input := fixture
			input.Cases = append([]PolicyRevisionCasePair(nil), fixture.Cases...)
			mutate(&input)
			report, err := ObserveGoooPolicyCounterexample(context.Background(), PolicyRevisionOperationContract(),
				"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", counterexampleExecutionJSON(t, input))
			if err == nil || report.Operation != nil || report.RevisionRequestJSON != "" {
				t.Fatalf("invalid input reached execution: %+v error=%v", report, err)
			}
			if name == "missing-pairs" && (report.Pending == nil || report.Pending.State != "UNKNOWN" ||
				report.Pending.Stage == "" || report.Pending.Step == "" || report.Pending.Reason == "" ||
				report.Pending.UnknownClass != "DIRECT_MISSING" || report.Pending.NextOperation == "" || report.Pending.BlockedBy == nil) {
				t.Fatal("missing pairs lost their six-field cause")
			}
		})
	}
	raw := counterexampleExecutionJSON(t, fixture)
	for name, bad := range map[string][]byte{
		"null": []byte("null"), "trailing": append(append([]byte(nil), raw...), []byte("{}")...),
		"unknown":   []byte(strings.TrimSuffix(string(raw), "}") + ",\"adopt\":true}"),
		"duplicate": []byte(strings.Replace(string(raw), "\"cases\":", "\"cases\":[],\"cases\":", 1)),
	} {
		t.Run(name, func(t *testing.T) {
			report, err := ObserveGoooPolicyCounterexample(context.Background(), PolicyRevisionOperationContract(),
				"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", bad)
			if err == nil || report.Schema != "" {
				t.Fatal("ambiguous input acquired a report")
			}
		})
	}
}

func TestCounterexampleExecutionPreservesStoppedAndFailedStages(t *testing.T) {
	source, input := counterexampleExecutionFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	raw := counterexampleExecutionJSON(t, input)
	unbound, err := ObserveGoooPolicyCounterexample(ctx, []byte("not an operation"),
		"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", raw)
	if err == nil || unbound.Schema != "" {
		t.Fatal("unbound operation was admitted")
	}
	failed, err := ObserveGoooPolicyCounterexample(ctx, PolicyRevisionOperationContract(),
		"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", raw)
	if err == nil || failed.Proposal.State != "PROPOSED" || failed.Operation == nil ||
		failed.Operation.NativeWorkerInvocations != 1 || failed.Operation.Observation == nil ||
		failed.Operation.Observation.ExecutionStatus != "FAILED" || len(failed.Operation.Observation.Pending) == 0 {
		t.Fatalf("failed execution lost its proposal or cause: %+v error=%v", failed, err)
	}
	input.Counterexample.Input.ValidatorExpectation = input.Counterexample.Observed.Decision
	stopped, err := ObserveGoooPolicyCounterexample(ctx, PolicyRevisionOperationContract(),
		"policy.gooo", source, "metapolicycompilation", "metapolicycompilation", counterexampleExecutionJSON(t, input))
	if err != nil || stopped.Proposal.State != "NOT_PROPOSED" || stopped.Operation != nil {
		t.Fatalf("declared agreement became an execution: %+v error=%v", stopped, err)
	}
}
