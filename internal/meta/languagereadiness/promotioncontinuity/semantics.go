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
