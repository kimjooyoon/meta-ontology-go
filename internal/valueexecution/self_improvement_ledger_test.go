package valueexecution

import "testing"

func ledgerTestEvidence(status SelfImprovementStatus, label string) ExecutedSelfImprovementObservation {
	candidateDigest := digestValue(label + "-candidate")
	return ExecutedSelfImprovementObservation{
		Schema:          ExecutedSelfImprovementObservationSchema,
		Candidate:       SelfImprovementObservation{Status: status, Digest: candidateDigest},
		CandidateDigest: candidateDigest,
		Status:          status,
		Reason:          "METRIC_" + string(status),
		NonAuthorizing:  true,
		Digest:          digestValue(label + "-execution"),
	}
}

func TestSelfImprovementEvidenceLedgerRequiresRepeatedImprovement(t *testing.T) {
	initial := NewSelfImprovementEvidenceLedger(2)
	one := initial.Append(ledgerTestEvidence(SelfImprovementStatusImproved, "one"))
	if len(initial.Observations) != 0 || len(one.Observations) != 1 {
		t.Fatalf("append must preserve prior ledger: initial=%d one=%d", len(initial.Observations), len(one.Observations))
	}
	observation := one.Observe()
	if observation.Decision != SelfImprovementLedgerDecisionInsufficientEvidence || observation.Improved != 1 {
		t.Fatalf("observation=%#v, want insufficient repeated evidence", observation)
	}
	eligible := one.Append(ledgerTestEvidence(SelfImprovementStatusImproved, "two")).Observe()
	if eligible.Decision != SelfImprovementLedgerDecisionPromotionEligible || eligible.Improved != 2 || !validDigest(eligible.Digest) {
		t.Fatalf("eligible=%#v, want two digest-backed improvements", eligible)
	}
}

func TestSelfImprovementEvidenceLedgerBlocksRegression(t *testing.T) {
	ledger := NewSelfImprovementEvidenceLedger(1).
		Append(ledgerTestEvidence(SelfImprovementStatusImproved, "improved")).
		Append(ledgerTestEvidence(SelfImprovementStatusRegressed, "regressed"))
	observation := ledger.Observe()
	if observation.Decision != SelfImprovementLedgerDecisionBlockedRegression || observation.Regressed != 1 || observation.Improved != 1 {
		t.Fatalf("observation=%#v, want preserved regression block", observation)
	}
}

func TestSelfImprovementEvidenceLedgerPreservesUnknown(t *testing.T) {
	ledger := NewSelfImprovementEvidenceLedger(1).
		Append(ledgerTestEvidence(SelfImprovementStatusUnknown, "unknown"))
	observation := ledger.Observe()
	if observation.Decision != SelfImprovementLedgerDecisionUnknown || observation.Unknown != 1 || observation.Reason != "LEDGER_CONTAINS_UNKNOWN_EVIDENCE" {
		t.Fatalf("observation=%#v, want UNKNOWN evidence", observation)
	}
}
