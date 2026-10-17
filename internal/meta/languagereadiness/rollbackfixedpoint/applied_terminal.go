package rollbackfixedpoint

// IsKnownAppliedTerminal recognizes a completed sandbox observation. The report
// still rejects recovery: its effects are nonzero and its original bytes, eight
// satisfied coordinates, and two rejected recovery coordinates stay unchanged.
func IsKnownAppliedTerminal(report Report) bool {
	return report.Decision == DecisionFailClosed && report.Reason == ReasonRejected &&
		report.Resolution == ResolutionExact && report.Mode == "" &&
		report.Summary.Satisfied == 8 && report.Summary.Total == 10 &&
		report.Summary.NotSatisfied == 2 && report.Summary.Unresolved == 0 &&
		report.Summary.RecoveredFixedPoints == 0 && report.Summary.AuthorizedPromotions == 0 &&
		!report.RepositoryMutationAuthorized && appliedTerminal(report.Source) && Validate(report) == nil
}

func appliedTerminal(source Source) bool {
	g, t := source.Guard, source.Transformation
	return source.CollectionError == "" && validSHA(source.ExpectedHeadSHA) &&
		g.HeadSHA == source.ExpectedHeadSHA && t.HeadSHA == source.ExpectedHeadSHA &&
		recoverable(g) && guardPopulationClosed(g) &&
		validDigest(g.FileSHA256) && validDigest(g.ReportDigest) &&
		validDigest(t.FileSHA256) && validDigest(t.LedgerDigest) &&
		t.Decision == "APPLIED" && t.Reason == "SANDBOX_EFFECTS_VERIFIED" &&
		t.WorkspaceMode == "DISPOSABLE_WORKTREE" && t.WriteBoundary == "SANDBOX_ONLY" &&
		t.OperationOutcome == "CLOSED" && t.ReceiptDecision == "CONFORMANT" &&
		t.Effects > 0 && t.AppliedEffects == t.Effects && t.ReceiptCount == t.Effects &&
		t.RefutedEffects == 0 && t.FailureCount == 0 && t.UnknownCount == 0 &&
		t.DirectUnknownCount == 0 && t.DependencyBlockedUnknownCount == 0 &&
		len(t.UnknownCausalDigest) == 64 && validDigest(t.UnknownCausalDigest) &&
		t.SourceWorkspaceUnchanged && source.RepositoryWrites == 0 && g.RepositoryWrites == 0 &&
		!g.RepositoryMutationAuthorized && !t.PromotionAuthorized
}

func guardPopulationClosed(g GuardEvidence) bool {
	return g.Total == 12 && g.Satisfied >= 0 && g.Satisfied <= g.Total &&
		g.NotSatisfied >= 0 && g.NotSatisfied <= g.Total &&
		g.Unresolved >= 0 && g.Unresolved <= g.Total &&
		g.Satisfied+g.NotSatisfied+g.Unresolved == g.Total
}
