package languagedeterministicquery

func stages(summary Summary) []StageReceipt {
	bindingsPass := summary.BindingPlans == 28
	replaysPass := summary.CanonicalReplays == 56 && summary.PermutationReplays == 28
	sealed := summary.CandidatePromotions == 0 && summary.UnknownAcceptances == 0 &&
		summary.GraphMutations == 0 && summary.EffectfulStages == 0
	return []StageReceipt{
		stage(1, "OBSERVE_CONCEPT_ARTIFACT", "FOUNDATION", "observe-explicit-concept-pass", summary.RegistryDrift == 0),
		stage(2, "REIFY_QUERY_PLAN", "FOUNDATION", "bind-versioned-query-plan-registry", summary.Executed == 32),
		stage(3, "PROJECT_META_GRAPH", "COHERENCE", "project-concept-bindings-to-prov", bindingsPass),
		stage(4, "NORMALIZE_REQUEST", "COHERENCE", "normalize-bounded-query-envelope", summary.CanonicalReplays >= 28),
		stage(5, "EXECUTE_QUERY", "COHERENCE", "execute-reified-deterministic-query-plan", bindingsPass),
		stage(6, "REPLAY_PERMUTATION", "COHERENCE", "replay-insertion-permutation", replaysPass),
		stage(7, "SEAL_EFFECTS", "REGRESSION", "seal-query-effects", sealed),
	}
}

func stage(ordinal int, name, proofChoice, operation string, passed bool) StageReceipt {
	status := "FAIL"
	if passed {
		status = "PASS"
	}
	return StageReceipt{
		Ordinal: ordinal, Stage: name, ProofChoice: proofChoice,
		MetaOperation: operation, Status: status, Effects: 0,
	}
}

func allStagesPassed(stages []StageReceipt) bool {
	if len(stages) != 7 {
		return false
	}
	for index, receipt := range stages {
		if receipt.Ordinal != index+1 || receipt.Status != "PASS" || receipt.Effects != 0 {
			return false
		}
	}
	return true
}
