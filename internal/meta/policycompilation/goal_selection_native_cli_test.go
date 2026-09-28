package policycompilation

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

type goalSelectionHarness struct {
	t         *testing.T
	ctx       context.Context
	directory string
}

func (h goalSelectionHarness) file(name string, data []byte) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	if err := os.WriteFile(path, data, 0400); err != nil {
		h.t.Fatal(err)
	}
	return path
}

func (h goalSelectionHarness) jsonFile(name string, value any) string {
	h.t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		h.t.Fatal(err)
	}
	return h.file(name, data)
}

func (h goalSelectionHarness) invoke(binary string, wantExit int, arguments ...string) []byte {
	h.t.Helper()
	command := exec.CommandContext(h.ctx, binary, arguments...)
	command.Env = append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if command.ProcessState == nil || command.ProcessState.ExitCode() != wantExit {
		h.t.Fatalf("public goal command: %v\n%s\n%s", err, stdout.Bytes(), stderr.Bytes())
	}
	return append([]byte(nil), stdout.Bytes()...)
}

func goalSelectionCase(policy CompiledPolicy, id, expected string) Case {
	return Case{
		ID: id, ValidatorExpectation: expected, EvidenceClass: "SYNTHETIC_FIXTURE",
		Provenance: "EXPLICIT_FROZEN_GOAL_CASE", ProducerAvailable: true, ConsumerAvailable: true,
		ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
		ObservedGeneratedJudgeDigest: DigestBytes(GenerateJudge(policy)), ObservedIndependentDigest: policy.SemanticDigest,
	}
}

func TestFrozenGoooGoalSelectionAndFollowingInvocation(t *testing.T) {
	goal, badRequest := revisionObservationFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	h := goalSelectionHarness{t: t, ctx: ctx, directory: t.TempDir()}
	_, location, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate checkout")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(location), "../../.."))
	witness := buildIndependentPublicCommand(t, ctx, root, h.directory, "meta-policy-compilation-witness")
	consumer := buildIndependentPublicCommand(t, ctx, root, h.directory, "meta-policy-compilation-consumer")
	next := filepath.Join(h.directory, "next-execution")
	if runtime.GOOS == "windows" {
		next += ".exe"
	}
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", next, "./cmd/meta-policy-compilation-witness/next-execution")
	build.Dir, build.Env = root, append(os.Environ(), "GOTOOLCHAIN=go1.27.0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build frozen-goal command: %v\n%s", err, output)
	}
	consumerBytes, err := os.ReadFile(consumer)
	if err != nil {
		t.Fatal(err)
	}
	goalPath := h.file("frozen-goal.gooo", goal)
	operationPath := h.file("operation.gooo", PolicyRevisionOperationContract())
	observe := func(name, sourcePath string, request PolicyRevisionObservationRequest) ([]byte, *PolicyRevisionObservation) {
		raw := h.invoke(witness, 0, "-policy", sourcePath, "-observe-revision", h.jsonFile(name+".json", request),
			"-revision-operation", operationPath, "-revision-consumer", consumer,
			"-revision-consumer-digest", DigestBytes(consumerBytes))
		envelope := independentPublicField[map[string]json.RawMessage](t, raw)
		operation := independentPublicField[PolicyRevisionOperationObservation](t, envelope["operation"])
		if operation.Observation == nil {
			t.Fatal("public predecessor omitted its source operation")
		}
		return raw, operation.Observation
	}
	badPredecessor, bad := observe("bad-revision", goalPath, badRequest)
	badPath := h.file("bad-policy.gooo", []byte(bad.CandidateSource))
	repair, err := ProposePolicyDecisionRevision("bad-policy.gooo", []byte(bad.CandidateSource),
		"metapolicycompilation", "metapolicycompilation", PolicyDecisionRevision{
			ExpectedSourceDigest: bad.CandidatePolicy.SourceDigest, Condition: ConditionSemanticEquivalence,
			FromDecision: DecisionFailClosed, ToDecision: DecisionPass,
		})
	if err != nil {
		t.Fatal(err)
	}
	missing := Case{ID: "repair-missing", ValidatorExpectation: DecisionUnknown,
		EvidenceClass: "SYNTHETIC_FIXTURE", Provenance: "EXPLICIT_MISSING_GOAL_CASE"}
	before := goalSelectionCase(bad.CandidatePolicy, "repair-fresh", DecisionFailClosed)
	after := goalSelectionCase(repair.Candidate, "repair-fresh", DecisionPass)
	stale := before
	stale.ID = "repair-stale"
	repairRequest := PolicyRevisionObservationRequest{
		ExpectedSourceDigest: bad.CandidatePolicy.SourceDigest, Condition: ConditionSemanticEquivalence,
		FromDecision: DecisionFailClosed, ToDecision: DecisionPass,
		Cases: []PolicyRevisionCasePair{{Baseline: before, Candidate: after}, {Baseline: stale, Candidate: stale}, {Baseline: missing, Candidate: missing}},
	}
	predecessor, repaired := observe("repair-revision", badPath, repairRequest)
	candidatePath := h.file("candidate.gooo", []byte(repaired.CandidateSource))
	predecessorPath := h.file("predecessor.json", predecessor)
	nextInputs := func(policy CompiledPolicy, expected string, previous []byte) NextPolicyExecutionRequest {
		fresh := goalSelectionCase(policy, "next-fresh", expected)
		old, upper := fresh, fresh
		old.ID, old.ObservedSourceDigest = "next-stale", DigestBytes([]byte("older-source"))
		upper.ID, upper.UpperDecision = "next-unknown-upper", "FIXED_POINT"
		absent := missing
		absent.ID = "next-missing"
		return NextPolicyExecutionRequest{
			Schema: NextPolicyExecutionRequestSchema, ExpectedSourceDigest: policy.SourceDigest,
			ExpectedSemanticDigest: policy.SemanticDigest, PredecessorDigest: DigestBytes(previous),
			Cases: []Case{fresh, old, absent, upper},
		}
	}
	input := nextInputs(repaired.CandidatePolicy, DecisionPass, predecessor)
	// The known rejection branches are fixed expectations, not the candidate's fresh decision.
	input.Cases[1].ValidatorExpectation, input.Cases[3].ValidatorExpectation = DecisionFailClosed, DecisionFailClosed
	requestPath := h.jsonFile("next-inputs.json", input)
	args := []string{"-policy", candidatePath, "-predecessor", predecessorPath, "-request", requestPath}
	selectionBytes := h.invoke(next, 0, append(append([]string(nil), args...),
		"-goal", goalPath, "-goal-digest", DigestBytes(goal), "-materialize-selection")...)
	selection := independentPublicField[GoalSelectionReport](t, selectionBytes)
	if selection.Decision != "SELECTED_FOR_FROZEN_GOOO_GOAL" || !selection.GoalSemanticMatch ||
		selection.Execution == nil || selection.Execution.ObservedCases != 4 || selection.Execution.ExpectationMismatches != 0 ||
		selection.SelectedSource != repaired.CandidateSource || selection.GeneralAdmission != "UNKNOWN" ||
		selection.GoalSourceDigest != DigestBytes(goal) || selection.MutationAuthority != 0 || selection.PromotionAuthority != 0 {
		t.Fatalf("goal selection was not source-bound: %s", selectionBytes)
	}
	directory := filepath.Dir(selection.MaterializedSource)
	if filepath.Base(selection.MaterializedSource) != "selected-policy.gooo" ||
		!strings.HasPrefix(filepath.Base(directory), "gooo-selected-policy-") ||
		filepath.Clean(filepath.Dir(directory)) != filepath.Clean(os.TempDir()) {
		t.Fatal("selection did not use its new temporary-output boundary")
	}
	defer os.RemoveAll(directory)
	materialized, err := os.ReadFile(selection.MaterializedSource)
	if err != nil || !bytes.Equal(materialized, []byte(repaired.CandidateSource)) {
		t.Fatal("materialized source does not contain the selected Gooo")
	}
	continuedBytes := h.invoke(next, 0, "-policy", selection.MaterializedSource,
		"-predecessor", predecessorPath, "-request", requestPath)
	continued := independentPublicField[NextPolicyExecutionReport](t, continuedBytes)
	if continued.Decision != "NEXT_EXECUTION_OBSERVED" || continued.SourceDigest != selection.SelectedSourceDigest ||
		continued.ObservedCases != 4 || continued.ExpectationMismatches != 0 {
		t.Fatal("following process did not consume the selected source")
	}
	badNext := nextInputs(bad.CandidatePolicy, DecisionFailClosed, badPredecessor)
	badArgs := []string{"-policy", badPath, "-predecessor", h.file("bad-predecessor.json", badPredecessor),
		"-request", h.jsonFile("relaxed-inputs.json", badNext)}
	selfExpectedBytes := h.invoke(next, 0, badArgs...)
	rejectedBytes := h.invoke(next, 1, append(append([]string(nil), badArgs...),
		"-goal", goalPath, "-goal-digest", DigestBytes(goal), "-materialize-selection")...)
	rejected := independentPublicField[GoalSelectionReport](t, rejectedBytes)
	if rejected.Reason != "CANDIDATE_DOES_NOT_MATCH_FROZEN_GOOO_GOAL" || rejected.Execution != nil || rejected.MaterializedSource != "" {
		t.Fatal("candidate-controlled expectations bypassed the frozen goal")
	}
	wrongPinBytes := h.invoke(next, 1, append(append([]string(nil), args...),
		"-goal", goalPath, "-goal-digest", DigestBytes([]byte("another-goal")))...)
	wrongPin := independentPublicField[GoalSelectionReport](t, wrongPinBytes)
	if wrongPin.Reason != "FROZEN_GOAL_DIGEST_MISMATCH" || wrongPin.Execution != nil {
		t.Fatal("a stale goal pin reached candidate execution")
	}
	invalidGoal := []byte("not a Gooo policy")
	unknownBytes := h.invoke(next, 2, append(append([]string(nil), args...),
		"-goal", h.file("invalid-goal.gooo", invalidGoal), "-goal-digest", DigestBytes(invalidGoal))...)
	unknown := independentPublicField[GoalSelectionReport](t, unknownBytes)
	if unknown.Pending == nil || unknown.Pending.Stage != "GOAL" || unknown.Pending.BlockedBy == nil ||
		unknown.Pending.NextOperation == "" || unknown.Execution != nil {
		t.Fatal("unavailable goal lost its causal boundary")
	}
	input.Cases[0].ValidatorExpectation = DecisionFailClosed
	contradictionBytes := h.invoke(next, 1, "-policy", candidatePath, "-predecessor", predecessorPath,
		"-request", h.jsonFile("contradictory-inputs.json", input), "-goal", goalPath, "-goal-digest", DigestBytes(goal))
	contradiction := independentPublicField[GoalSelectionReport](t, contradictionBytes)
	if !contradiction.GoalSemanticMatch || contradiction.Execution == nil || contradiction.Execution.ExpectationMismatches != 1 ||
		contradiction.MaterializedSource != "" || contradiction.SelectedSource != "" {
		t.Fatal("a matching goal erased the observed counterexample")
	}
	actualGoal, err := os.ReadFile(goalPath)
	if err != nil || !bytes.Equal(actualGoal, goal) {
		t.Fatal("the fixed Gooo goal was modified")
	}
	event, err := json.Marshal(map[string]any{
		"goal_source": string(goal), "goal_digest": DigestBytes(goal),
		"predecessor_stdout": string(predecessor), "predecessor_digest": DigestBytes(predecessor),
		"selection_stdout": string(selectionBytes), "selection_stdout_digest": DigestBytes(selectionBytes),
		"continued_stdout": string(continuedBytes), "continued_stdout_digest": DigestBytes(continuedBytes),
		"self_expected_stdout": string(selfExpectedBytes), "rejected": rejected, "wrong_goal_pin": wrongPin,
		"unknown_goal": unknown, "execution_counterexample": contradiction,
		"evidence_class": "SYNTHETIC_FIXTURE", "general_admission": "UNKNOWN", "improvement": "UNKNOWN",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("FROZEN_GOOO_GOAL_SELECTION_WITNESS=%s", event)
}
