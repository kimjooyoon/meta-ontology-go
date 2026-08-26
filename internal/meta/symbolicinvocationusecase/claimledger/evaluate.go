package claimledger

func projectEvidence(report *Report, claim Claim, spec ClaimSpec, sources map[string]sourceState, subject string) {
	report.Metrics.InScopeClaimTotal++
	evidenceID := "evidence:" + spec.ID
	claim.EvidenceRefs = []string{evidenceID}
	evidence, observed, found, unknownReason := selectEvidence(spec, sources, evidenceID)
	if !found {
		claim.Status, claim.Truth, claim.Reason = "UNKNOWN", "UNDETERMINED", unknownReason
		report.Metrics.UnknownTotal++
		report.OpenClaimIDs = append(report.OpenClaimIDs, spec.ID)
		report.Evidence = append(report.Evidence, evidence)
		eventType := "EVIDENCE_MISSING"
		if evidence.Status == "REJECTED" {
			eventType = "EVIDENCE_REJECTED"
		}
		addEvent(report, spec, eventType, evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_UNKNOWN", claim.Status, claim.Reason, evidenceID)
		report.Claims = append(report.Claims, claim)
		return
	}
	matched, expectedDigest := evidenceMatches(*spec.Evidence, observed, subject)
	evidence.ObservedValueDigest = digestValue(observed)
	evidence.ExpectedValueDigest = expectedDigest
	if matched {
		evidence.Status = "VERIFIED"
		claim.Status, claim.Truth, claim.Reason = "DISCHARGED", "SATISFIED", "EVIDENCE_VERIFIED"
		report.Metrics.DischargedTotal++
		addEvent(report, spec, "EVIDENCE_VERIFIED", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_DISCHARGED", claim.Status, claim.Reason, evidenceID)
	} else {
		evidence.Status = "FAILED"
		claim.Status, claim.Truth, claim.Reason = "REFUTED", "VIOLATED", spec.RefutedReason
		report.Metrics.RefutedTotal++
		addEvent(report, spec, "EVIDENCE_REJECTED", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_REFUTED", claim.Status, claim.Reason, evidenceID)
	}
	report.Evidence = append(report.Evidence, evidence)
	report.Claims = append(report.Claims, claim)
}
