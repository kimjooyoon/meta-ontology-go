package main

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const extractionContractSourceMismatch = "EXTRACTION_CONTRACT_SOURCE_MISMATCH"

var errExtractorContractSourceMismatch = errors.New(
	extractionContractSourceMismatch + ": prepare a new extraction with the current contract")

func extractionReportDecodeError(err error) *operationError {
	reason, next := "INSTANCE_EVIDENCE_MALFORMED", "report-counterexample"
	if errors.Is(err, errExtractorContractSourceMismatch) {
		reason, next = extractionContractSourceMismatch, "restore-extraction-plan"
	}
	return newOperationError("evaluate-operation", "decode-function-extraction-report", reason, "KNOWN_CONTRADICTION", next)
}

// Classify a rejected observation with a different, well-formed contract source
// identity, including when the contract meaning also changed. This supplies a
// next step without admitting or rewriting the old observation.
func extractorContractSourceMismatch(report extractorReport) bool {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return false
	}
	for _, subject := range report.Subjects {
		for _, evidence := range subject.Evidence {
			if validBareDigest(evidence.ContractSourceDigest) && validBareDigest(evidence.ContractSemanticDigest) &&
				evidence.ContractSourceDigest != contract.SourceDigest {
				return true
			}
		}
	}
	return false
}
