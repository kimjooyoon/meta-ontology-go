package policycompilation

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestPinnedGoooGoalDerivesCounterexampleAndContinues(t *testing.T) {
	if os.Getenv("CI") != "true" {
		t.Skip("native Go 1.27 Actions witness")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	dir := t.TempDir()
	goal, counterexample, _ := counterexampleRevisionFixture(t)
	broken, err := ProposePolicyDecisionRevision("goal.gooo", goal, "metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(goal), Condition: counterexample.Observed.MatchedCondition,
		FromDecision: DecisionPass, ToDecision: DecisionFailClosed,
	})
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(broken.CandidateSource)
	request := PolicyGoalCounterexampleRequest{
		ExpectedSourceDigest: DigestBytes(source), ExpectedGoalDigest: DigestBytes(goal),
		SourceCase: goalCounterexampleCase(counterexample.Input, broken.Candidate),
		GoalCase:   goalCounterexampleCase(counterexample.Input, broken.Original),
	}
	rawRequest, _ := json.Marshal(request)
	sourcePath := goalCounterexampleFile(t, dir, "source.gooo", source)
	goalPath := goalCounterexampleFile(t, dir, "goal.gooo", goal)
	requestPath := goalCounterexampleFile(t, dir, "request.json", rawRequest)
	command := buildGoalCounterexampleCommand(t, ctx, dir, "derive", "./cmd/meta-policy-compilation-witness/goal-counterexample")
	raw := runGoalCounterexampleCommand(t, ctx, command, "-policy", sourcePath, "-goal", goalPath, "-request", requestPath)
	var observed PolicyGoalCounterexampleObservation
	if err := json.Unmarshal(raw, &observed); err != nil {
		t.Fatal(err)
	}
	if observed.State != "PROPOSED" || observed.GeneratedBatches != 2 || len(observed.SourceResults) != 1 ||
		len(observed.GoalResults) != 1 || observed.SourceResults[0].Decision != DecisionFailClosed ||
		observed.GoalResults[0].Decision != DecisionPass || observed.Proposal == nil || observed.Proposal.Revision == nil {
		t.Fatalf("goal did not produce an executable counterexample: %s", raw)
	}
	proposal := observed.Proposal
	if proposal.Revision.Condition != observed.SourceResults[0].MatchedCondition ||
		proposal.Revision.FromDecision != observed.SourceResults[0].Decision ||
		proposal.Revision.ToDecision != observed.GoalResults[0].Decision || len(proposal.DerivedFields) != 3 ||
		proposal.Counterexample.Input.ValidatorExpectation != DecisionPass || !reflect.DeepEqual(*observed.Request, request) {
		t.Fatal("revision fields were not derived from the goal and original execution")
	}
	negative := map[string]PolicyGoalCounterexampleObservation{}
	for _, name := range []string{"caller-expectation", "stale-source", "wrong-goal-pin", "different-case", "goal-unknown", "agreement"} {
		input, originalSource := request, source
		switch name {
		case "caller-expectation":
			input.SourceCase.ValidatorExpectation = DecisionPass
		case "stale-source":
			input.SourceCase.ObservedSourceDigest = DigestBytes([]byte("stale"))
		case "wrong-goal-pin":
			input.ExpectedGoalDigest = DigestBytes([]byte("other goal"))
		case "different-case":
			input.GoalCase.ID = "unrelated"
		case "goal-unknown":
			input.SourceCase.ProducerAvailable, input.GoalCase.ProducerAvailable = false, false
		case "agreement":
			originalSource, input.ExpectedSourceDigest, input.SourceCase = goal, DigestBytes(goal), input.GoalCase
		}
		inputRaw, _ := json.Marshal(input)
		result := ObserveGoalPolicyCounterexample(ctx, sourcePath, originalSource, goal, inputRaw, "metapolicycompilation", "metapolicycompilation")
		negative[name] = result
		expected := map[string]string{"caller-expectation": "REFUTED", "stale-source": "UNKNOWN", "wrong-goal-pin": "REFUTED", "different-case": "REFUTED", "goal-unknown": "UNKNOWN", "agreement": "NOT_PROPOSED"}[name]
		if result.State != expected || (result.Proposal != nil && result.Proposal.CandidateSource != "") {
			t.Fatalf("%s manufactured a candidate: %+v", name, result)
		}
		if result.State == "UNKNOWN" && (result.Pending == nil || result.Pending.State != "UNKNOWN" ||
			result.Pending.Stage == "" || result.Pending.Step == "" || result.Pending.Reason == "" ||
			result.Pending.UnknownClass == "" || result.Pending.NextOperation == "" || result.Pending.BlockedBy == nil) {
			t.Fatalf("%s lost UNKNOWN causality", name)
		}
	}
	for _, malformed := range []string{"null", "{}", string(rawRequest) + "{}"} {
		rejected := ObserveGoalPolicyCounterexample(ctx, sourcePath, source, goal, []byte(malformed), "metapolicycompilation", "metapolicycompilation")
		if rejected.State != "REFUTED" || rejected.GeneratedBatches != 0 || rejected.Proposal != nil {
			t.Fatal("malformed request executed a policy")
		}
	}
	goalCounterexampleContinue(t, ctx, dir, sourcePath, goalPath, goal, request, proposal, raw, negative)
}

func goalCounterexampleContinue(t *testing.T, ctx context.Context, dir, sourcePath, goalPath string, goal []byte, request PolicyGoalCounterexampleRequest, proposal *PolicyCounterexampleProposal, raw []byte, negative map[string]PolicyGoalCounterexampleObservation) {
	t.Helper()
	baseline := request.SourceCase
	baseline.ValidatorExpectation = proposal.Revision.FromDecision
	candidate := goalCounterexampleCase(request.SourceCase, *proposal.CandidatePolicy)
	candidate.ValidatorExpectation = proposal.Revision.ToDecision
	revision := PolicyRevisionObservationRequest{
		ExpectedSourceDigest: proposal.Revision.ExpectedSourceDigest, Condition: proposal.Revision.Condition,
		FromDecision: proposal.Revision.FromDecision, ToDecision: proposal.Revision.ToDecision,
		Cases: []PolicyRevisionCasePair{{Baseline: baseline, Candidate: candidate}},
	}
	revisionRaw, _ := json.Marshal(revision)
	revisionPath := goalCounterexampleFile(t, dir, "revision.json", revisionRaw)
	operationPath := goalCounterexampleFile(t, dir, "operation.gooo", PolicyRevisionOperationContract())
	consumer := buildGoalCounterexampleCommand(t, ctx, dir, "consumer", "./cmd/meta-policy-compilation-consumer")
	consumerBytes, err := os.ReadFile(consumer)
	if err != nil {
		t.Fatal(err)
	}
	witness := buildGoalCounterexampleCommand(t, ctx, dir, "witness", "./cmd/meta-policy-compilation-witness")
	predecessor := runGoalCounterexampleCommand(t, ctx, witness, "-policy", sourcePath, "-observe-revision", revisionPath,
		"-revision-operation", operationPath, "-revision-consumer", consumer, "-revision-consumer-digest", DigestBytes(consumerBytes))
	predecessorPath := goalCounterexampleFile(t, dir, "predecessor.json", predecessor)
	candidatePath := goalCounterexampleFile(t, dir, "candidate.gooo", []byte(proposal.CandidateSource))
	nextRequest, _ := json.Marshal(map[string]any{
		"schema": "gooo/meta-policy-next-execution-request/v1", "expected_source_digest": DigestBytes([]byte(proposal.CandidateSource)),
		"expected_semantic_digest": proposal.CandidatePolicy.SemanticDigest, "predecessor_digest": DigestBytes(predecessor),
		"cases": []Case{candidate},
	})
	nextPath := goalCounterexampleFile(t, dir, "next.json", nextRequest)
	next := buildGoalCounterexampleCommand(t, ctx, dir, "next", "./cmd/meta-policy-compilation-witness/next-execution")
	selection := runGoalCounterexampleCommand(t, ctx, next, "-policy", candidatePath, "-predecessor", predecessorPath,
		"-request", nextPath, "-goal", goalPath, "-goal-digest", DigestBytes(goal), "-materialize-selection")
	var selected struct {
		Decision           string `json:"decision"`
		MaterializedSource string `json:"materialized_source"`
	}
	if err := json.Unmarshal(selection, &selected); err != nil || selected.Decision != "SELECTED_FOR_FROZEN_GOOO_GOAL" || selected.MaterializedSource == "" {
		t.Fatalf("goal-derived candidate was not selected: %s (%v)", selection, err)
	}
	t.Cleanup(func() { os.RemoveAll(filepath.Dir(selected.MaterializedSource)) })
	continued := runGoalCounterexampleCommand(t, ctx, next, "-policy", selected.MaterializedSource, "-predecessor", predecessorPath, "-request", nextPath)
	var result struct {
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal(continued, &result); err != nil || result.Decision != "NEXT_EXECUTION_OBSERVED" {
		t.Fatalf("following process did not consume the selection: %s (%v)", continued, err)
	}
	record, _ := json.Marshal(map[string]any{
		"scope": "SYNTHETIC_FIXTURE", "general_admission": "UNKNOWN", "improvement": "UNKNOWN",
		"goal_source": string(goal), "declared_request": request, "counterexample_stdout": string(raw),
		"counterexample_stdout_digest": DigestBytes(raw), "derived_revision_request": revision,
		"predecessor_stdout": string(predecessor), "predecessor_digest": DigestBytes(predecessor),
		"selection_stdout": string(selection), "selection_digest": DigestBytes(selection),
		"continued_stdout": string(continued), "continued_digest": DigestBytes(continued), "negative_cases": negative,
	})
	t.Logf("GOAL_DERIVED_COUNTEREXAMPLE_NATIVE_WITNESS=%s", record)
}

func goalCounterexampleCase(input Case, policy CompiledPolicy) Case {
	input.ValidatorExpectation = ""
	input.ObservedSourceDigest, input.ObservedArtifactSourceDigest = policy.SourceDigest, policy.SourceDigest
	input.ObservedGeneratedJudgeDigest, input.ObservedIndependentDigest = DigestBytes(GenerateJudge(policy)), policy.SemanticDigest
	return input
}

func goalCounterexampleFile(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildGoalCounterexampleCommand(t *testing.T, ctx context.Context, dir, name, pkg string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", path, pkg)
	command.Dir = filepath.Join("..", "..", "..")
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build native command: %v: %s", err, output)
	}
	return path
}

func runGoalCounterexampleCommand(t *testing.T, ctx context.Context, binary string, args ...string) []byte {
	t.Helper()
	command := exec.CommandContext(ctx, binary, args...)
	output, err := command.Output()
	if err != nil {
		t.Fatalf("native command failed: %v (%s)", err, output)
	}
	return output
}
