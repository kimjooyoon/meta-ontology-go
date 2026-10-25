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

// Classify a rejected observation whose contract meaning still matches, while
// its source identity differs. This supplies a next step without admitting it.
func extractorContractSourceMismatch(report extractorReport) bool {
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		return false
	}
	for _, subject := range report.Subjects {
		for _, evidence := range subject.Evidence {
			if evidence.ContractSourceDigest != "" && evidence.ContractSourceDigest != contract.SourceDigest &&
				evidence.ContractSemanticDigest == contract.SemanticDigest {
				return true
			}
		}
	}
	return false
}
