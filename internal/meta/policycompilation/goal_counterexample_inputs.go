package policycompilation

import "strings"

func prepareGoalCounterexample(report PolicyGoalCounterexampleObservation, filename string, source, goal, raw []byte, pkg, namespace string) ([2]CompiledPolicy, PolicyGoalCounterexampleObservation, bool) {
	var policies [2]CompiledPolicy
	var request *PolicyGoalCounterexampleRequest
	if err := decodeStrictJSON(raw, &request); err != nil || request == nil {
		if err != nil {
			report.Diagnostic = err.Error()
		}
		return policies, stopGoalCounterexample(report, "REFUTED", "INPUT", "GOAL_COUNTEREXAMPLE_REQUEST_INVALID", ""), false
	}
	report.Request = request
	if !ValidDigest(request.ExpectedSourceDigest) || request.ExpectedSourceDigest != report.SourceDigest ||
		!ValidDigest(request.ExpectedGoalDigest) || request.ExpectedGoalDigest != report.GoalDigest {
		return policies, stopGoalCounterexample(report, "REFUTED", "INPUT", "SOURCE_OR_GOAL_PIN_MISMATCH", ""), false
	}
	a, b := request.SourceCase, request.GoalCase
	if a.ValidatorExpectation != "" || b.ValidatorExpectation != "" {
		return policies, stopGoalCounterexample(report, "REFUTED", "INPUT", "CALLER_EXPECTATION_NOT_ALLOWED", ""), false
	}
	if strings.TrimSpace(a.ID) == "" || a.ID != b.ID || a.UpperDecision != b.UpperDecision ||
		a.ProducerAvailable != b.ProducerAvailable || a.ConsumerAvailable != b.ConsumerAvailable ||
		a.EvidenceClass != b.EvidenceClass || a.Provenance != b.Provenance {
		return policies, stopGoalCounterexample(report, "REFUTED", "INPUT", "SOURCE_AND_GOAL_CASE_SCOPE_MISMATCH", ""), false
	}
	for i, item := range []struct {
		name string
		source []byte
		input Case
	}{{filename, source, a}, {filename + ".goal.gooo", goal, b}} {
		policy, err := CompileForIdentity(item.name, item.source, pkg, namespace)
		if err != nil {
			report.Diagnostic = err.Error()
			return policies, stopGoalCounterexample(report, "UNKNOWN", "INPUT", "SOURCE_OR_GOAL_NOT_COMPILED", "AMBIGUOUS"), false
		}
		if item.input.ObservedSourceDigest != policy.SourceDigest ||
			item.input.ObservedArtifactSourceDigest != policy.SourceDigest ||
			item.input.ObservedGeneratedJudgeDigest != DigestBytes(GenerateJudge(policy)) ||
			item.input.ObservedIndependentDigest != policy.SemanticDigest {
			return policies, stopGoalCounterexample(report, "UNKNOWN", "INPUT", "SOURCE_OR_GOAL_CASE_SNAPSHOT_STALE", "STALE"), false
		}
		policies[i] = policy
	}
	if policies[0].PolicyID != policies[1].PolicyID {
		return policies, stopGoalCounterexample(report, "REFUTED", "GOAL", "GOAL_POLICY_ID_MISMATCH", ""), false
	}
	report.GoalSemanticDigest = policies[1].SemanticDigest
	return policies, report, true
}

func stopGoalCounterexample(report PolicyGoalCounterexampleObservation, state, stage, reason, class string) PolicyGoalCounterexampleObservation {
	report.State, report.Reason = state, reason
	if state == "UNKNOWN" {
		pending := revisionPending(stage, "OBSERVE_GOAL_BOUND_COUNTEREXAMPLE", reason, "RESOLVE_"+stage+"_AND_REPEAT")
		pending.UnknownClass = class
		if class == "DEPENDENCY_BLOCKED" {
			pending.BlockedBy = []string{strings.ToLower(stage) + ".result"}
		}
		report.Pending = &pending
	}
	return report
}
