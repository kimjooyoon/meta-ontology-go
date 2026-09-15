package main

func finalizeClaimDependencies(report claimDependencyReport) claimDependencyReport {
	cycle := claimDependencyCycleResidue(report.Nodes, report.Edges)
	report.Summary.CyclicActivities = len(cycle)
	if len(cycle) > 0 {
		return refuteClaimDependencies(report, "CAUSALITY", "SELECT_PROOF_STRUCTURE", "CLAIM_DEPENDENCY_CYCLE_DETECTED", "SELECT_FOUNDATION_OR_BREAK_CYCLE", claimDependencyRegression, cycle)
	}
	if report.Summary.RecoverableRoots == 0 {
		return unknownClaimDependencies(report, "FOUNDATION", "SELECT_RECOVERABLE_ROOT", "RECOVERABLE_ROOT_UNAVAILABLE", "DIRECT_MISSING", "DECLARE_RECOVERABLE_ROOT", claimDependencyFoundation, []string{})
	}
	report.Decision = claimDependencyObserved
	report.Resolution = claimDependencyResolution{
		State: claimStateClosed, Reason: "CLAIM_DEPENDENCY_CAUSALITY_OBSERVED", NextOperation: claimResolutionNone,
		ProofChoice: claimDependencyCoherence, BlockedBy: []string{},
	}
	report.Indicators = buildClaimDependencyIndicators(report)
	return report
}
