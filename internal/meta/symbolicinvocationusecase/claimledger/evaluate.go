package claimledger

import "strings"

func projectEvidence(report *Report, claim Claim, spec ClaimSpec, observation map[string]any, subject string) {
	report.Metrics.InScopeClaimTotal++
	evidenceID := "evidence:" + spec.ID
	claim.EvidenceRefs = []string{evidenceID}
	path, observed, found := lookupAny(observation, spec.Evidence.Paths)
	evidence := Evidence{
		ID: evidenceID, ClaimID: spec.ID,
		SourcePath: strings.Join(spec.Evidence.Paths, "|"), SourceDigest: report.ObservationDigest,
	}
	if !found {
		evidence.Status = "MISSING"
		claim.Status, claim.Truth, claim.Reason = "UNKNOWN", "UNDETERMINED", spec.UnknownReason
		report.Metrics.UnknownTotal++
		report.OpenClaimIDs = append(report.OpenClaimIDs, spec.ID)
		report.Evidence = append(report.Evidence, evidence)
		addEvent(report, spec, "EVIDENCE_MISSING", evidence.Status, claim.Reason, evidenceID)
		addEvent(report, spec, "CLAIM_UNKNOWN", claim.Status, claim.Reason, evidenceID)
		report.Claims = append(report.Claims, claim)
		return
	}
	evidence.SourcePath = path
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
