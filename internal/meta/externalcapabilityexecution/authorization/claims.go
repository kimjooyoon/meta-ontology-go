package authorization

func makeClaims(indicators []Indicator) ([]Claim, []UnknownEvidence, int, int, int) {
	claims := make([]Claim, 0, len(indicators))
	unknowns := []UnknownEvidence{}
	open, discharged, rejected := 0, 0, 0
	for index, indicator := range indicators {
		claim := Claim{ClaimID: "gooo.claim.external-capability-authorization-" +
			metricSpecs[index].ID + ".v1", Stage: indicator.Stage,
			Statement: metricSpecs[index].Claim, EvidenceDigest: indicator.EvidenceDigest}
		switch indicator.Status {
		case StatusSatisfied:
			claim.Status = "DISCHARGED"
			discharged++
		case StatusUnsatisfied:
			claim.Status = "REJECTED"
			rejected++
		default:
			claim.Status = "OPEN"
			claim.EvidenceDigest = ""
			open++
			unknowns = append(unknowns, UnknownEvidence{Stage: indicator.Stage,
				IndicatorID: indicator.MetricID, Reason: indicator.UnknownReason})
		}
		claims = append(claims, claim)
	}
	return claims, unknowns, open, discharged, rejected
}
