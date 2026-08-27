package languageresourcebudget

func buildTransitions(semantic Semantic, resourceComplete, resourceMissing bool, violations int, writeSetTo, writeSetReason string) []ClaimTransition {
	semanticTo, semanticReason := semantic.ClaimState, semantic.Reason
	if semanticTo == "" {
		semanticTo, semanticReason = "REFUTED", "SEMANTIC_PRODUCER_EVIDENCE_INVALID"
	}
	resourceTo, resourceReason := "DISCHARGED", "RESOURCE_ENVELOPE_OBSERVED"
	if resourceMissing {
		resourceTo, resourceReason = "OPEN", "RESOURCE_SAMPLE_MISSING"
	} else if !resourceComplete {
		resourceTo, resourceReason = "OPEN", "RESOURCE_SAMPLE_INVALID"
	} else if violations > 0 {
		resourceTo, resourceReason = "REFUTED", "RESOURCE_BUDGET_EXCEEDED"
	}
	return []ClaimTransition{
		{Sequence: 1, ClaimID: "semantic-meaning", From: "OPEN", To: semanticTo, Stage: "REDUCE", Step: "semantic-verdict", Reason: semanticReason},
		{Sequence: 2, ClaimID: "runner-resource-envelope", From: "OPEN", To: resourceTo, Stage: "REDUCE", Step: "resource-verdict", Reason: resourceReason},
		{Sequence: 3, ClaimID: "net-repository-state", From: "OPEN", To: writeSetTo, Stage: "REDUCE", Step: "effect-verdict", Reason: writeSetReason},
	}
}
