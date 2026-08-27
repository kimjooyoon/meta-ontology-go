package valuecatalog

func buildOperationClaims(report Report) []Claim {
	checks := operationSpecChecks(report)
	claims := make([]Claim, 0, len(checks))
	for _, check := range checks {
		status, evidence := "OPEN", ""
		if check.satisfied {
			status, evidence = "DISCHARGED", check.evidence
		}
		claims = append(claims, Claim{
			ClaimID: "gooo.claim.operation-spec." + check.id + ".v1",
			Stage: check.stage, Statement: check.statement, Status: status, EvidenceDigest: evidence,
		})
	}
	return claims
}

func closeOperationSpec(report Report) Report {
	report.Claims = buildOperationClaims(report)
	verified, discharged := 0, 0
	for index, claim := range report.Claims {
		if operationSpecChecks(report)[index].satisfied {
			verified++
		}
		if claim.Status == "DISCHARGED" {
			discharged++
		}
	}
	report.OperationSpecMetrics = OperationSpecMetrics{
		MetricID: OperationSpecMetricID, FixedAxisTotal: OperationSpecAxisTotal,
		VerifiedTotal: verified, CoverageBasisPoints: verified * 10_000 / OperationSpecAxisTotal,
		UnknownPathCount: boolInt(report.Decision == DecisionFailClosed),
		OpenClaims: len(report.Claims) - discharged, DischargedClaims: discharged,
	}
	return report
}
