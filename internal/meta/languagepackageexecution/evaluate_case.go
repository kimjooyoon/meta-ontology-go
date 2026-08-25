package languagepackageexecution

import "github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/packageexecution"

func evaluateCase(spec CaseSpec, evidence CaseEvidence) CaseResult {
	receipt := evidence.Receipt
	result := CaseResult{ID: spec.ID, Decision: receipt.Decision, Reason: receipt.Reason, Resolution: receipt.Resolution, ReceiptDigest: receipt.Digest}
	if evidence.ID == "" {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "CASE_EVIDENCE_MISSING"
		result.Resolution = "EXACT"
		return result
	}
	if receipt.Decision != "PASS" && receipt.Decision != "FAIL_CLOSED" {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "PACKAGE_EXECUTION_DECISION_UNKNOWN"
		result.Resolution = "LOWER_RESOLUTION"
		return result
	}
	if packageexecution.Validate(receipt) != nil {
		result.Decision = "FAIL_CLOSED"
		result.Reason = "CASE_RECEIPT_INVALID"
		result.Resolution = "EXACT"
		return result
	}
	result.Satisfied = receipt.Decision == spec.ExpectedDecision && receipt.Reason == spec.ExpectedReason
	return result
}
