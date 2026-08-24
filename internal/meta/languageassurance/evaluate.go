package languageassurance

import "sort"

func Evaluate(subjectSHA string, transaction Transaction) (Report, error) {
	if err := validateInput(subjectSHA, transaction); err != nil {
		return Report{}, err
	}
	definitions := Denominator()
	operations := CanonicalMetaOperations()
	obligations, operating := observeObligations(definitions)
	findings := append(detectSelfMinting(transaction), detectRoleConflicts(transaction)...)
	findings = append(findings, detectUnknownLaundering(transaction)...)
	sort.Slice(findings, func(i, j int) bool {
		return findings[i].MetricID+findings[i].PathID < findings[j].MetricID+findings[j].PathID
	})

	authorityObserved := len(transaction.AuthorityRoutes) > 0
	rolesObserved := len(transaction.RoleBindings) > 0
	decisionsObserved := len(transaction.DecisionTransitions) > 0
	evidenceObserved := boolInt(authorityObserved) + boolInt(rolesObserved) + boolInt(decisionsObserved)
	selfMinting := countFindings(findings, MetricSelfMinting)
	roleConflicts := countFindings(findings, MetricRoleConflict)
	unknownLaundering := countFindings(findings, MetricUnknownLaundering)
	unknownTop := countUnknownTop(transaction.DecisionTransitions)

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
		RepositoryWrites:          0,
	}
	summary.UnresolvedIndicators = unresolved(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths)
	summary.ViolatedGuardrails = positive(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths)
	decision, reason, resolution := decide(summary)
	indicators := buildIndicators(summary)
	report := Report{
		Schema: ReportSchema, SubjectSHA: subjectSHA,
		TransactionDigest: digest(transaction), DenominatorID: DenominatorID,
		DenominatorDigest: digest(definitions), AssuranceDecision: AssurancePartial,
		CandidateDecision: decision, CandidateReason: reason, CandidateResolution: resolution,
		Denominator: definitions, Obligations: obligations, MetaOperations: operations,
		RoleConflictPairs: RoleConflictPairs(), UnknownLaunderingOutputs: UnknownLaunderingOutputs(),
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
