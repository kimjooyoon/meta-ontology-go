package languageassurance

import "testing"

func withTestReceipt(t *testing.T, transaction Transaction) Transaction {
	t.Helper()
	report, err := Evaluate(testSHA, transaction)
	if err != nil {
		t.Fatal(err)
	}
	summary := report.Summary
	summary.UnresolvedIndicators = unresolved(summary.SelfMintingPaths, summary.RoleConflictPaths, summary.UnknownLaunderingPaths, summary.ExactSnapshotBindingBPS)
	transaction.RawReconstructions = []RawReconstructionReceipt{expectedRawReceipt(testSHA, transaction, summary, candidateFor(summary))}
	return transaction
}

func TestRawReconstructionResolution(t *testing.T) {
	missing := evaluateForTest(t, independentTransaction())
	if missing.CandidateDecision != CandidateFailClosed || missing.CandidateReason != ReasonEvidenceUnknown || missing.Summary.RawReconstructionBPS != nil {
		t.Fatalf("missing reconstruction was not fail-closed: %+v", missing.Summary)
	}
	transaction := withTestReceipt(t, independentTransaction())
	transaction.RawReconstructions[0].Observation.CandidateDecision = CandidateBlock
	mismatch := evaluateForTest(t, transaction)
	if mismatch.CandidateDecision != CandidateBlock || mismatch.CandidateReason != ReasonRawMismatch || metricValue(mismatch.Summary.RawReconstructionBPS) != 0 || metricValue(mismatch.Summary.RawReconstructionMismatchPaths) != 1 {
		t.Fatalf("reconstruction mismatch was not blocked: %+v", mismatch.Summary)
	}
}
