package languageassurance

func Evaluate(subjectSHA string, transaction Transaction) (Report, error) {
	if err := validateInput(subjectSHA, transaction); err != nil {
		return Report{}, err
	}
	definitions := Denominator()
	operations := CanonicalMetaOperations()
	obligations, operating := observeObligations(definitions)
	findings := append(detectSelfMinting(transaction), detectRoleConflicts(transaction)...)
	findings = append(findings, detectUnknownLaundering(transaction)...)
	findings = append(findings, detectSnapshotMismatches(subjectSHA, transaction)...)
	authorityObserved := len(transaction.AuthorityRoutes) > 0
	rolesObserved := len(transaction.RoleBindings) > 0
	decisionsObserved := len(transaction.DecisionTransitions) > 0
	evidenceObserved := boolInt(authorityObserved) + boolInt(rolesObserved) + boolInt(decisionsObserved)
	selfMinting := countFindings(findings, MetricSelfMinting)
	roleConflicts := countFindings(findings, MetricRoleConflict)
	unknownLaundering := countFindings(findings, MetricUnknownLaundering)
	unknownTop := countUnknownTop(transaction.DecisionTransitions)
	snapshotMismatches := countFindings(findings, MetricSnapshotBinding)
	snapshotBPS, snapshotPaths := observeSnapshotBindings(transaction.SnapshotBindings, snapshotMismatches)

	summary := Summary{
		DenominatorTotal:          len(definitions),
		Operating:                 operating,
		NotImplemented:            len(definitions) - operating,
		ImplementationCoverageBPS: operating * 10000 / len(definitions),
		EvidenceGroupsObserved:    evidenceObserved,
		EvidenceGroupsTotal:       3,
		EvidenceCoverageBPS:       evidenceObserved * 10000 / 3,
		SelfMintingPaths:          observedValue(authorityObserved, selfMinting),
		RoleConflictPaths:         observedValue(rolesObserved, roleConflicts),
		UnknownLaunderingPaths:    observedValue(decisionsObserved, unknownLaundering),
		UnknownTopDecisions:       observedValue(decisionsObserved, unknownTop),
		SnapshotBindingsObserved:  len(transaction.SnapshotBindings), SnapshotBindingsRequired: len(snapshotEvidenceIDs),
		ExactSnapshotBindingBPS: snapshotBPS, SnapshotMismatchPaths: snapshotPaths,
		RepositoryWrites: 0,
	}
	summary.ViolatedGuardrails = positive(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths)
	summary, findings, candidate := completeReconstruction(subjectSHA, transaction, summary, findings)
	indicators := buildIndicators(summary)
	report := Report{
		Schema: ReportSchema, SubjectSHA: subjectSHA,
		TransactionDigest: digest(transaction), DenominatorID: DenominatorID,
		DenominatorDigest: digest(definitions), AssuranceDecision: AssurancePartial,
		CandidateDecision: candidate.Decision, CandidateReason: candidate.Reason, CandidateResolution: candidate.Resolution,
		Denominator: definitions, Obligations: obligations, MetaOperations: operations,
		RoleConflictPairs: RoleConflictPairs(), UnknownLaunderingOutputs: UnknownLaunderingOutputs(), SnapshotEvidenceIDs: SnapshotEvidenceIDs(),
		Transaction: transaction, Findings: findings, Indicators: indicators, Summary: summary,
	}
	seal(&report)
	return report, nil
}

func observeObligations(definitions []ObligationDefinition) ([]ObligationObservation, int) {
	observations := make([]ObligationObservation, 0, len(definitions))
	operating := 0
	for _, definition := range definitions {
		observation := ObligationObservation{MetricID: definition.MetricID, Status: "NOT_IMPLEMENTED", Resolution: ResolutionNone}
		if operation, ok := operatingOperations[definition.MetricID]; ok {
			observation.Status, observation.Resolution, observation.MetaOperation = "OPERATING", ResolutionExact, operation
			operating++
		}
		observations = append(observations, observation)
	}
	return observations, operating
}
