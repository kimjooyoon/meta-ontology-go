package languagesourceexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"

func receiptCase(spec CaseSpec, raw []byte, receipt sourceexecution.Receipt, err error) CaseResult {
	result := caseResult(spec, "UNKNOWN", "SOURCE_EXECUTION_RECEIPT_UNKNOWN", raw)
	if err != nil {
		result.Reason = err.Error()
		return result
	}
	result.Status, result.Reason = "NOT_SATISFIED", "SOURCE_EXECUTION_CASE_MISMATCH"
	if receipt.Decision == spec.ExpectedDecision && receipt.Reason == spec.ExpectedReason {
		result.Status, result.Reason = "SATISFIED", "SOURCE_EXECUTION_CASE_EXACT"
	}
	return result
}

func replayCase(spec CaseSpec, firstRaw, replayRaw []byte, firstErr, replayErr error) CaseResult {
	result := caseResult(spec, "UNKNOWN", "SOURCE_EXECUTION_REPLAY_UNKNOWN", replayRaw)
	if firstErr != nil || replayErr != nil {
		return result
	}
	result.Status, result.Reason = "NOT_SATISFIED", "SOURCE_EXECUTION_REPLAY_MISMATCH"
	if string(firstRaw) == string(replayRaw) {
		result.Status, result.Reason = "SATISFIED", "SOURCE_EXECUTION_REPLAY_EXACT"
	}
	return result
}

func caseResult(spec CaseSpec, status, reason string, raw []byte) CaseResult {
	return CaseResult{ID: spec.ID, Status: status, Reason: reason,
		ProofChoice: spec.ProofChoice, MetaOperation: spec.MetaOperation, EvidenceDigest: digestBytes(raw)}
}
