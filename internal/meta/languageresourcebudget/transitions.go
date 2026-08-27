package languageresourcebudget

func buildTransitions(semantic Semantic, complete bool, violations int, semanticErr bool) []ClaimTransition {
	semanticTo, semanticReason := "DISCHARGED", "SEMANTIC_ARTIFACT_REPLAY_STABLE"
	if semanticErr {
		semanticTo, semanticReason = "REFUTED", semantic.Reason
	}
	resourceTo, resourceReason := "DISCHARGED", "RESOURCE_ENVELOPE_OBSERVED"
	if !complete {
		resourceTo, resourceReason = "OPEN", "RESOURCE_SAMPLE_MISSING"
	}
	if violations > 0 {
		resourceTo, resourceReason = "REFUTED", "RESOURCE_BUDGET_EXCEEDED"
	}
	return []ClaimTransition{
		{Sequence: 1, ClaimID: "semantic-meaning", From: "OPEN", To: semanticTo, Stage: "REDUCE", Step: "semantic-verdict", Reason: semanticReason},
		{Sequence: 2, ClaimID: "runner-resource-envelope", From: "OPEN", To: resourceTo, Stage: "REDUCE", Step: "resource-verdict", Reason: resourceReason},
		{Sequence: 3, ClaimID: "read-only-observation", From: "OPEN", To: "DISCHARGED", Stage: "REDUCE", Step: "effect-verdict", Reason: "EFFECT_BOUNDARY_VERIFIED"},
	}
}
