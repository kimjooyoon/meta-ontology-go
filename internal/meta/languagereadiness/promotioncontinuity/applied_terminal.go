package promotioncontinuity

func knownAppliedRecovery(head string, guard GuardEvidence, t RecoveryEvidence) bool {
	return canonicalGuard(guard) && canonicalRecovery(t) && t.HeadSHA == head &&
		t.Decision == DecisionFailClosed && t.Reason == "ROLLBACK_FIXED_POINT_REJECTED" &&
		t.Resolution == "EXACT" && t.Mode == "" &&
		t.GuardDecision == DecisionFailClosed && t.GuardReason == "GUARDED_PROMOTION_EVIDENCE_UNKNOWN" &&
		t.GuardResolution == "LOWER_RESOLUTION" && t.GuardSatisfied >= 0 && t.GuardTotal == 12 &&
		t.GuardSatisfied < t.GuardTotal &&
		t.GuardUnresolved > 0 && t.GuardSatisfied+t.GuardUnresolved == t.GuardTotal &&
		t.GuardRepositoryWrites == 0 && !t.GuardMutationAuthorized &&
		t.Satisfied == 8 && t.Total == 10 && t.Unresolved == 0 && t.ReadinessBPS == 8000 &&
		t.RecoveredFixedPoints == 0 && t.AuthorizedPromotions == 0 &&
		t.TransformationHeadSHA == head && t.TransformationDecision == "APPLIED" &&
		t.TransformationReason == "SANDBOX_EFFECTS_VERIFIED" &&
		t.TransformationWorkspaceMode == "DISPOSABLE_WORKTREE" &&
		t.TransformationOperationOutcome == "CLOSED" && t.TransformationReceiptDecision == "CONFORMANT" &&
		t.TransformationEffects > 0 && t.TransformationAppliedEffects == t.TransformationEffects &&
		t.TransformationReceiptCount == t.TransformationEffects &&
		t.TransformationRefutedEffects == 0 && t.TransformationFailureCount == 0 &&
		t.TransformationUnknownCount == 0 && t.TransformationDirectUnknownCount == 0 &&
		t.TransformationDependencyBlockedUnknownCount == 0 &&
		validCausalDigest(t.TransformationUnknownCausalDigest) &&
		validDigest(t.TransformationCausalBindingDigest) && t.TransformationCausalBindingDigest == causalBindingDigest(t) &&
		t.WriteBoundary == "SANDBOX_ONLY" && t.SourceWorkspaceUnchanged &&
		!t.TransformationAuthorization && zeroRecoveryWrites(t) && !t.MutationAuthorized &&
		guard.HeadSHA == head && guard.Decision == DecisionFailClosed &&
		guard.Reason == "GUARDED_PROMOTION_EVIDENCE_UNKNOWN" && guard.Resolution == "LOWER_RESOLUTION" &&
		guard.Satisfied == t.GuardSatisfied && guard.Total == t.GuardTotal && guard.Unresolved == t.GuardUnresolved &&
		guard.RepositoryWrites == 0 && !guard.PromotionAuthorized && !guard.MutationAuthorized
}
