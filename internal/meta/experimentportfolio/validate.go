package experimentportfolio

import (
	"encoding/hex"
	"strings"
)

var validStatuses = map[string]bool{
	"OPEN":       true,
	"DISCHARGED": true,
	"REFUTED":    true,
}

func inputReason(input Input) string {
	if reason := contractReason(input.Contract); reason != "" {
		return reason
	}
	if !validSubjectSHA(input.SubjectSHA) {
		return "PORTFOLIO_SUBJECT_SHA_INVALID"
	}
	if len(input.Receipts) != ExpectedCandidates {
		return "PORTFOLIO_RECEIPT_COUNT_INVALID"
	}
	seen := map[string]bool{}
	for _, receipt := range input.Receipts {
		if seen[receipt.CandidateID] {
			return "PORTFOLIO_CANDIDATE_DUPLICATE"
		}
		seen[receipt.CandidateID] = true
		if reason := receiptReason(input.Contract, input.SubjectSHA, receipt); reason != "" {
			return reason
		}
	}
	for _, candidate := range input.Contract.Candidates {
		if !seen[candidate.ID] {
			return "PORTFOLIO_CANDIDATE_MISSING"
		}
	}
	return ""
}

func receiptReason(contract Contract, subjectSHA string, receipt Receipt) string {
	candidate, ok := candidateContract(contract, receipt.CandidateID)
	if !ok {
		return "PORTFOLIO_CANDIDATE_UNKNOWN"
	}
	if receipt.Schema != contract.ReceiptSchema || receipt.SubjectSHA != subjectSHA ||
		receipt.SourcePath != candidate.SourcePath || receipt.Producer != receiptProducer ||
		receipt.Consumer != receiptConsumer || receipt.MetaOperation != candidate.MetaOperation ||
		receipt.ProofChoice != candidate.ProofChoice {
		return "PORTFOLIO_RECEIPT_IDENTITY_MISMATCH"
	}
	if !validDigest(receipt.SourceDigest) || !validDigest(receipt.FactsDigest) || receipt.Digest != receiptDigest(receipt) {
		return "PORTFOLIO_RECEIPT_DIGEST_INVALID"
	}
	if len(receipt.CoordinateVector) != len(contract.CoordinateIDs) {
		return "PORTFOLIO_COORDINATE_VECTOR_LENGTH_INVALID"
	}
	for index, coordinate := range receipt.CoordinateVector {
		if reason := coordinateReason(contract, candidate, coordinate, index); reason != "" {
			return reason
		}
	}
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return "PORTFOLIO_RECEIPT_NOT_READ_ONLY"
	}
	if len(receipt.Counterexamples) > contract.CounterexampleSlots || !uniqueCounterexamples(receipt.Counterexamples) {
		return "PORTFOLIO_COUNTEREXAMPLES_INVALID"
	}
	if len(receipt.UnknownLocations) > contract.UnknownLocationSlots || !uniqueUnknownLocations(receipt.UnknownLocations) {
		return "PORTFOLIO_UNKNOWN_LOCATIONS_INVALID"
	}
	if reason := evidenceReason(receipt); reason != "" {
		return reason
	}
	return ""
}

func coordinateReason(contract Contract, candidate CandidateContract, coordinate Coordinate, index int) string {
	if coordinate.ID != contract.CoordinateIDs[index] || coordinate.Denominator != contract.CoordinateDenominators[coordinate.ID] ||
		coordinate.Producer != receiptProducer || coordinate.Consumer != receiptConsumer ||
		coordinate.MetaOperation != candidate.MetaOperation || coordinate.ProofChoice != candidate.ProofChoice ||
		coordinate.Stage == "" || coordinate.Step == "" || coordinate.Reason == "" || !validStatuses[coordinate.Status] {
		return "PORTFOLIO_COORDINATE_IDENTITY_INVALID"
	}
	if coordinate.Numerator < 0 || coordinate.Numerator > coordinate.Denominator {
		return "PORTFOLIO_COORDINATE_BOUNDS_INVALID"
	}
	switch coordinate.Status {
	case "DISCHARGED":
		if coordinate.Numerator != coordinate.Denominator {
			return "PORTFOLIO_DISCHARGE_WITHOUT_FULL_NUMERATOR"
		}
	case "REFUTED":
		if coordinate.Numerator == coordinate.Denominator {
			return "PORTFOLIO_REFUTATION_WITH_FULL_NUMERATOR"
		}
	}
	return ""
}

func evidenceReason(receipt Receipt) string {
	coordinates := make(map[string]Coordinate, len(receipt.CoordinateVector))
	for _, coordinate := range receipt.CoordinateVector {
		coordinates[coordinate.ID] = coordinate
	}
	if coordinates["counterexample-boundary"].Numerator != len(receipt.Counterexamples) {
		return "PORTFOLIO_COUNTEREXAMPLE_COUNT_MISMATCH"
	}
	if coordinates["unknown-localization"].Numerator != len(receipt.UnknownLocations) {
		return "PORTFOLIO_UNKNOWN_LOCATION_COUNT_MISMATCH"
	}
	dischargedExtensions := 0
	for _, evidence := range receipt.ExtensionEvidence {
		if evidence.ID == "" || evidence.Claim == "" || evidence.Evidence == "" || evidence.Stage == "" || evidence.Step == "" || evidence.Reason == "" || !validStatuses[evidence.Status] {
			return "PORTFOLIO_EXTENSION_EVIDENCE_INVALID"
		}
		if evidence.Status == "DISCHARGED" {
			dischargedExtensions++
		}
	}
	if coordinates["extension-evidence"].Numerator != dischargedExtensions {
		return "PORTFOLIO_EXTENSION_COUNT_MISMATCH"
	}
	return ""
}

func uniqueCounterexamples(values []Counterexample) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value.ID == "" || value.Location == "" || value.Claim == "" || value.Stage == "" || value.Step == "" || value.Reason == "" || seen[value.ID] {
			return false
		}
		seen[value.ID] = true
	}
	return true
}

func uniqueUnknownLocations(values []UnknownLocation) bool {
	seen := map[string]bool{}
	for _, value := range values {
		if value.ID == "" || value.Path == "" || value.Stage == "" || value.Step == "" || value.Reason == "" || seen[value.ID] {
			return false
		}
		seen[value.ID] = true
	}
	return true
}

func validSubjectSHA(value string) bool {
	if len(value) != 40 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func validDigest(value string) bool {
	return len(value) == 71 && strings.HasPrefix(value, "sha256:")
}
