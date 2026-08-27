package languageproofartifactverifier

func evaluateInterventions(inputs []InterventionInput, head, phase string) []InterventionResult {
	result := make([]InterventionResult, 0, len(inputs))
	for _, input := range inputs {
		before := verifyArtifact(input.Before.Artifact, input.Before.Source, input.Before.Operation, input.Before.Recipe, head, phase)
		after := verifyArtifact(input.After.Artifact, input.After.Source, input.After.Operation, input.After.Recipe, head, phase)
		beforePolicy, afterPolicy := before, after
		if len(input.PolicyBefore.Artifact) > 0 && len(input.PolicyAfter.Artifact) > 0 {
			beforePolicy = verifyArtifact(input.PolicyBefore.Artifact, input.PolicyBefore.Source, input.PolicyBefore.Operation, input.PolicyBefore.Recipe, head, phase)
			afterPolicy = verifyArtifact(input.PolicyAfter.Artifact, input.PolicyAfter.Source, input.PolicyAfter.Operation, input.PolicyAfter.Recipe, head, phase)
		}
		item := InterventionResult{ID: input.ID, Kind: input.Kind,
			RawSourceDigestBefore: before.SourceDigest, RawSourceDigestAfter: after.SourceDigest,
			SemanticDigestBefore: before.SemanticDigest, SemanticDigestAfter: after.SemanticDigest,
			OperationReceiptDigestBefore: before.OperationDigest, OperationReceiptDigestAfter: after.OperationDigest,
			EvidenceLinkDigestBefore: before.EvidenceLinkDigest, EvidenceLinkDigestAfter: after.EvidenceLinkDigest,
			ClaimTransitionDigestBefore: before.ClaimTransitionDigest, ClaimTransitionDigestAfter: after.ClaimTransitionDigest,
			ConsumerDecisionBefore: before.Decision, ConsumerDecisionAfter: after.Decision,
			PolicySemanticDigestBefore: beforePolicy.Policy.SemanticDigest, PolicySemanticDigestAfter: afterPolicy.Policy.SemanticDigest,
			PolicySelectionBefore: beforePolicy.Policy.SelectedIssue, PolicySelectionAfter: afterPolicy.Policy.SelectedIssue,
			PolicySelectionRankBefore: beforePolicy.Policy.SelectedRank, PolicySelectionRankAfter: afterPolicy.Policy.SelectedRank,
			PolicyMembershipDigestBefore: beforePolicy.Policy.ObservedIssueMembershipDigest, PolicyMembershipDigestAfter: afterPolicy.Policy.ObservedIssueMembershipDigest,
			PolicyObservedIssueCountBefore: beforePolicy.Policy.ObservedIssueCount, PolicyObservedIssueCountAfter: afterPolicy.Policy.ObservedIssueCount}
		item.RawDigestChanged = nonEmptyDifferent(item.RawSourceDigestBefore, item.RawSourceDigestAfter)
		item.SemanticDigestChanged = nonEmptyDifferent(item.SemanticDigestBefore, item.SemanticDigestAfter)
		item.OperationReceiptChanged = nonEmptyDifferent(item.OperationReceiptDigestBefore, item.OperationReceiptDigestAfter)
		item.EvidenceLinksChanged = nonEmptyDifferent(item.EvidenceLinkDigestBefore, item.EvidenceLinkDigestAfter)
		item.ClaimTransitionsChanged = nonEmptyDifferent(item.ClaimTransitionDigestBefore, item.ClaimTransitionDigestAfter)
		item.SemanticDigestPreserved = equalNonEmpty(item.SemanticDigestBefore, item.SemanticDigestAfter)
		item.ConsumerDecisionPreserved = equalNonEmpty(item.ConsumerDecisionBefore, item.ConsumerDecisionAfter)
		item.PolicySelectionChanged = nonEmptyDifferent(item.PolicySemanticDigestBefore, item.PolicySemanticDigestAfter) &&
			(item.PolicySelectionBefore != item.PolicySelectionAfter || item.PolicySelectionRankBefore != item.PolicySelectionRankAfter)
		item.PolicySelectionPreserved = equalNonEmpty(item.PolicySemanticDigestBefore, item.PolicySemanticDigestAfter) &&
			item.PolicySelectionBefore != "" && item.PolicySelectionBefore == item.PolicySelectionAfter && item.PolicySelectionRankBefore == item.PolicySelectionRankAfter
		item.PolicyMembershipPreserved = equalNonEmpty(item.PolicyMembershipDigestBefore, item.PolicyMembershipDigestAfter) &&
			item.PolicyObservedIssueCountBefore == item.PolicyObservedIssueCountAfter
		item.Status, item.Reason = "NOT_SATISFIED", "INTERVENTION_CONTRACT_VIOLATED"
		switch input.Kind {
		case "SEMANTIC":
			if before.Decision == "PASS" && after.Decision == "PASS" && item.RawDigestChanged && item.SemanticDigestChanged &&
				item.OperationReceiptChanged && item.EvidenceLinksChanged && item.ClaimTransitionsChanged && item.ConsumerDecisionPreserved && item.PolicyMembershipPreserved && item.PolicySelectionChanged {
				item.Status, item.Reason = "SATISFIED", "SEMANTIC_INTERVENTION_CHANGED_OPERATION_EVIDENCE_AND_TRANSITION"
			}
		case "NONSEMANTIC":
			if before.Decision == "PASS" && after.Decision == "PASS" && item.RawDigestChanged && item.SemanticDigestPreserved && item.ConsumerDecisionPreserved {
				if item.PolicyMembershipPreserved && item.PolicySelectionPreserved {
					item.Status, item.Reason = "SATISFIED", "COMMENT_ONLY_INTERVENTION_PRESERVED_SEMANTICS"
				}
			}
		}
		result = append(result, item)
	}
	return result
}

func nonEmptyDifferent(left, right string) bool {
	return left != "" && right != "" && left != right
}

func equalNonEmpty(left, right string) bool {
	return left != "" && left == right
}
