package valuecatalog

func buildOperationClaims(checks []operationSpecCheck) []Claim {
	claims := make([]Claim, 0, len(checks))
	for _, check := range checks {
		status, evidence := ClaimStatusOpen, ""
		if check.satisfied {
			status, evidence = ClaimStatusDischarged, check.evidence
		}
		claims = append(claims, Claim{
			ClaimID: "gooo.claim.operation-spec." + check.id + ".v1",
			Stage:   check.stage, Statement: check.statement, Status: status, EvidenceDigest: evidence,
		})
	}
	return claims
}

func closeOperationSpec(report Report) Report {
	checks := operationSpecChecks(report)
	report.Claims = buildOperationClaims(checks)
	report.ClaimTransitions = buildOperationClaimTransitions(report, checks)
	report.ClaimTransitionHead = report.ClaimTransitions[len(report.ClaimTransitions)-1].TransitionDigest
	verified, discharged := 0, 0
	for index, claim := range report.Claims {
		if checks[index].satisfied {
			verified++
		}
		if claim.Status == ClaimStatusDischarged {
			discharged++
		}
	}
	registered, accepted, unavailable := countClaimTransitionEvents(report.ClaimTransitions)
	report.OperationSpecMetrics = OperationSpecMetrics{
		MetricID: OperationSpecMetricID, FixedAxisTotal: OperationSpecAxisTotal,
		VerifiedTotal: verified, CoverageBasisPoints: verified * 10_000 / OperationSpecAxisTotal,
		UnknownPathCount: boolInt(report.Decision == DecisionFailClosed),
		OpenClaims: len(report.Claims) - discharged, DischargedClaims: discharged,
		TransitionEventTotal: len(report.ClaimTransitions), RegistrationEventTotal: registered,
		EvidenceAcceptedTotal: accepted, EvidenceUnavailableTotal: unavailable,
	}
	return report
}
