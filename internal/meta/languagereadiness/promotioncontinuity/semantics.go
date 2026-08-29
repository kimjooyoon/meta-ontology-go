package promotioncontinuity

func canonicalGuard(guard GuardEvidence) bool {
	return guard.Schema == guardSchema && validSHA(guard.HeadSHA) &&
		validDigest(guard.FileSHA256) && validDigest(guard.ReportDigest)
}

func authorizedGuard(guard GuardEvidence, head string) bool {
	return canonicalGuard(guard) && guard.HeadSHA == head &&
		guard.Decision == "AUTHORIZED" && guard.Reason == "MERGED_PUSH_PROMOTION_AUTHORIZED" &&
		guard.Resolution == "EXACT" && guard.Satisfied == 12 && guard.Total == 12 &&
		guard.Unresolved == 0 && guard.RepositoryWrites == 0 &&
		guard.PromotionAuthorized && !guard.MutationAuthorized
}

func canonicalRecovery(recovery RecoveryEvidence) bool {
	return recovery.Schema == recoverySchema && validSHA(recovery.HeadSHA) &&
		validDigest(recovery.FileSHA256) && validDigest(recovery.ReportDigest)
}

func authorizedRecovery(recovery RecoveryEvidence, head string) bool {
	return canonicalRecovery(recovery) && recovery.HeadSHA == head &&
		recovery.Decision == "PASS" && recovery.Resolution == "EXACT" &&
		recovery.Mode == "PROMOTION_AUTHORIZED" && recovery.GuardDecision == "AUTHORIZED" &&
		recovery.GuardResolution == "EXACT" && recovery.Satisfied == 10 && recovery.Total == 10 &&
		recovery.Unresolved == 0 && recovery.ReadinessBPS == 10000 &&
		recovery.RecoveredFixedPoints == 0 && recovery.AuthorizedPromotions == 1 &&
		effectBoundary(recovery) && zeroRecoveryWrites(recovery) && !recovery.MutationAuthorized
}

func effectBoundary(recovery RecoveryEvidence) bool {
	return canonicalRecovery(recovery) && recovery.TransformationDecision == "FIXED_POINT" &&
		recovery.TransformationEffects == 0 && recovery.WriteBoundary == "SANDBOX_ONLY" &&
		recovery.SourceWorkspaceUnchanged && !recovery.TransformationAuthorization
}

func zeroRecoveryWrites(recovery RecoveryEvidence) bool {
	return recovery.SourceRepositoryWrites == 0 && recovery.SummaryRepositoryWrites == 0 &&
		recovery.RepositoryWrites == 0
}

func authorityBoundary(guard GuardEvidence, recovery RecoveryEvidence) bool {
	return canonicalGuard(guard) && canonicalRecovery(recovery) && guard.RepositoryWrites == 0 &&
		zeroRecoveryWrites(recovery) && !guard.MutationAuthorized && !recovery.MutationAuthorized
}

func knownMixedRecovery(head string, guard GuardEvidence, recovery RecoveryEvidence) bool {
	t := recovery
	return canonicalGuard(guard) && canonicalRecovery(t) && t.HeadSHA == head &&
		t.Decision == DecisionFailClosed && t.Reason == "MIXED_REFUTED_NON_PROMOTABLE" &&
		t.Resolution == "EXACT" && t.Mode == "MIXED_REFUTED_NON_PROMOTABLE" &&
		t.GuardDecision == DecisionFailClosed && t.GuardReason == "GUARDED_PROMOTION_EVIDENCE_UNKNOWN" &&
		t.GuardResolution == "LOWER_RESOLUTION" && t.GuardSatisfied > 0 &&
		t.GuardTotal > t.GuardSatisfied && t.GuardSatisfied+t.GuardUnresolved == t.GuardTotal &&
		t.GuardRepositoryWrites == 0 && !t.GuardMutationAuthorized &&
		t.Satisfied == 8 && t.Total == 10 && t.Unresolved == 0 &&
		t.ReadinessBPS == 8000 && t.RecoveredFixedPoints == 0 &&
		t.AuthorizedPromotions == 0 && t.TransformationHeadSHA == head &&
		t.TransformationDecision == "APPLIED" &&
		t.TransformationReason == "SANDBOX_EFFECTS_VERIFIED" &&
		t.TransformationWorkspaceMode == "DISPOSABLE_WORKTREE" &&
		t.TransformationEffects == 2 && t.TransformationAppliedEffects == 1 &&
		t.TransformationRefutedEffects == 1 &&
		t.TransformationOperationOutcome == "MIXED_CLOSED_REFUTED" &&
		t.TransformationReceiptDecision == "REFUTED" &&
		t.TransformationReceiptCount == 1 && t.TransformationFailureCount == 1 &&
		t.TransformationUnknownCount == 5 && t.TransformationDirectUnknownCount == 0 &&
		t.TransformationDependencyBlockedUnknownCount == 5 &&
		validCausalDigest(t.TransformationUnknownCausalDigest) &&
		validDigest(t.TransformationCausalBindingDigest) &&
		t.TransformationCausalBindingDigest == causalBindingDigest(t) &&
		t.WriteBoundary == "SANDBOX_ONLY" && t.SourceWorkspaceUnchanged &&
		!t.TransformationAuthorization && t.SourceRepositoryWrites == 0 &&
		t.SummaryRepositoryWrites == 0 && t.RepositoryWrites == 0 &&
		!t.MutationAuthorized && guard.HeadSHA == head &&
		guard.Decision == DecisionFailClosed && guard.Reason == "GUARDED_PROMOTION_EVIDENCE_UNKNOWN" &&
		guard.Resolution == "LOWER_RESOLUTION" && guard.Satisfied == t.GuardSatisfied &&
		guard.Total == t.GuardTotal && guard.Unresolved == t.GuardUnresolved &&
		guard.RepositoryWrites == 0 && !guard.PromotionAuthorized && !guard.MutationAuthorized
}

func IsKnownNonPromotingTerminal(report Report) bool {
	return report.Decision == DecisionFailClosed && report.Reason == ReasonMixed &&
		report.Resolution == "EXACT" && report.Mode == ModeMixed &&
		report.MetaOperation == OperationMixed &&
		knownMixedRecovery(report.Source.ExpectedHeadSHA, report.Source.Guard, report.Source.Recovery)
}
