package valueexecution

import "testing"

func testEnvironmentInputs() EnvironmentInputs {
	return EnvironmentInputs{
		SourceDigest:        digestBytes([]byte("source")),
		SemanticFingerprint: digestBytes([]byte("semantic")),
		ToolchainDigest:     digestBytes([]byte("toolchain")),
		ContractDigest:      digestBytes([]byte("contract")),
		ModelDigest:         digestBytes([]byte("model")),
		SkillDigest:         digestBytes([]byte("skill")),
		GatewayPolicyDigest: digestBytes([]byte("gateway")),
	}
}

func testReplayExecution() Execution {
	execution := Execution{
		Scope:       "test-scope",
		PlanDigest:  digestBytes([]byte("plan")),
		InputDigest: digestBytes([]byte("input")),
		Phase:       ExecutionPhaseCompleted,
	}
	execution.ExecutionDigest = executionDigest(execution)
	return execution
}

func TestObserveEnvironmentRequiresEveryIdentity(t *testing.T) {
	ready := ObserveEnvironment(testEnvironmentInputs())
	if ready.Status != EnvironmentStatusReady || !ready.NonAuthorizing || !validDigest(ready.Digest) {
		t.Fatalf("expected a complete non-authorizing environment observation, got %#v", ready)
	}
	if transition := CompareEnvironments(ready, ready); transition != EnvironmentTransitionUnchanged {
		t.Fatalf("expected unchanged environment, got %s", transition)
	}
	changedInputs := testEnvironmentInputs()
	changedInputs.ToolchainDigest = digestBytes([]byte("other-toolchain"))
	changed := ObserveEnvironment(changedInputs)
	if transition := CompareEnvironments(ready, changed); transition != EnvironmentTransitionChanged {
		t.Fatalf("expected changed environment, got %s", transition)
	}
	missing := testEnvironmentInputs()
	missing.ModelDigest = ""
	unknown := ObserveEnvironment(missing)
	if unknown.Status != EnvironmentStatusUnknown || len(unknown.Missing) != 1 || unknown.Missing[0] != "model_digest" {
		t.Fatalf("expected model identity to remain missing, got %#v", unknown)
	}
}

func TestCompareReplayWithEnvironmentFailsClosedOnChange(t *testing.T) {
	baseline := testReplayExecution()
	candidate := testReplayExecution()
	baselineEnvironment := ObserveEnvironment(testEnvironmentInputs())
	changedInputs := testEnvironmentInputs()
	changedInputs.SkillDigest = digestBytes([]byte("other-skill"))
	changedEnvironment := ObserveEnvironment(changedInputs)

	changed := CompareReplayWithEnvironment(baseline, candidate, baselineEnvironment, changedEnvironment)
	if changed.State != ReplayUnknown || changed.Reason != "REPLAY_ENVIRONMENT_CHANGED" || changed.SameEnvironment {
		t.Fatalf("expected changed environment to block replay, got %#v", changed)
	}

	stable := CompareReplayWithEnvironment(baseline, candidate, baselineEnvironment, baselineEnvironment)
	if stable.State != ReplayClosed || !stable.SameEnvironment || stable.Replay.State != ReplayClosed {
		t.Fatalf("expected stable environment to permit deterministic replay, got %#v", stable)
	}
}
