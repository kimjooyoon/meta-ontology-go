package promotioncontinuity

func Build(input Input) (Report, error) {
	guard, err := readGuard(input.GuardPath)
	if err != nil {
		return Report{}, err
	}
	recovery, err := readRecovery(input.RecoveryPath)
	if err != nil {
		return Report{}, err
	}
	return Evaluate(input.ExpectedHeadSHA, guard, recovery), nil
}

func Evaluate(head string, guard GuardEvidence, recovery RecoveryEvidence) Report {
	headExact := validSHA(head)
	guardExact := canonicalGuard(guard)
	recoveryExact := canonicalRecovery(recovery)
	subjectExact := headExact && guard.HeadSHA == head && recovery.HeadSHA == head
	guardAuthorized := authorizedGuard(guard, head)
	recoveryAuthorized := authorizedRecovery(recovery, head)
	effectsExact := effectBoundary(recovery)
	authorityExact := authorityBoundary(guard, recovery)
	coordinates := []Coordinate{
		coordinate("expected-subject", "FOUNDATION", headExact),
		coordinate("canonical-guard", "FOUNDATION", guardExact),
		coordinate("canonical-recovery", "FOUNDATION", recoveryExact),
		coordinate("exact-subject-link", "COHERENCE", subjectExact),
		coordinate("guard-authorized", "COHERENCE", guardAuthorized),
		coordinate("recovery-authorized", "COHERENCE", recoveryAuthorized),
		coordinate("zero-effect-boundary", "REGRESSION", effectsExact),
		coordinate("zero-authority-boundary", "REGRESSION", authorityExact),
	}
	satisfied := 0
	for _, item := range coordinates {
		if item.Status == "SATISFIED" {
			satisfied++
		}
	}
	decision, reason, resolution := "FAIL_CLOSED", "PROMOTION_CONTINUITY_UNKNOWN", "LOWER_RESOLUTION"
	if satisfied == len(coordinates) {
		decision, reason, resolution = "PASS", "PROMOTION_AUTHORIZED_CONTINUITY_PROVEN", "EXACT"
	}
	writes := guard.RepositoryWrites + recovery.SourceRepositoryWrites +
		recovery.SummaryRepositoryWrites + recovery.RepositoryWrites
	report := Report{
		Schema: Schema, Decision: decision, Reason: reason, Resolution: resolution,
		Producer: "promotioncontinuity.Evaluate", Consumer: "self-improvement-cycle",
		MetaOperation: "prove-authorized-successor",
		Source: Source{ExpectedHeadSHA: head, Guard: guard, Recovery: recovery},
		Summary: Summary{Satisfied: satisfied, Total: len(coordinates),
			Unresolved: len(coordinates) - satisfied, ReadinessBPS: satisfied * 10000 / len(coordinates),
			AuthorizedGuardReceipts: boolInt(guardAuthorized),
			AuthorizedRecoveryRoutes: boolInt(recoveryAuthorized), RepositoryWrites: writes},
		Coordinates: coordinates, RepositoryWrites: writes,
		RepositoryMutationAuthorized: guard.MutationAuthorized || recovery.MutationAuthorized,
	}
	report.Indicators = buildIndicators(report, guardAuthorized, recoveryAuthorized, effectsExact, authorityExact)
	report.Proofs = buildProofs(report, headExact && guardExact && recoveryExact,
		subjectExact && guardAuthorized && recoveryAuthorized, effectsExact && authorityExact)
	return seal(report)
}

func coordinate(id, choice string, passed bool) Coordinate {
	if passed {
		return Coordinate{ID: id, ProofChoice: choice, Status: "SATISFIED", Reason: "COORDINATE_EXACTLY_PROVEN"}
	}
	return Coordinate{ID: id, ProofChoice: choice, Status: "UNRESOLVED", Reason: "EVIDENCE_NOT_EXACT"}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
