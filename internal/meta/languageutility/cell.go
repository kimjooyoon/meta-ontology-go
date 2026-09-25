package languageutility

import (
	"encoding/hex"
	"strings"
)

func classifyCell(useCase UseCaseSpec, stage StageSpec, observed *CellObservation, duplicate bool, graph GraphObservation) CellResult {
	result := CellResult{UseCaseID: useCase.ID, StageID: stage.ID, ProofChoice: stage.ProofChoice}
	if duplicate {
		return unknownResult(result, "INDEX_EVIDENCE", "DUPLICATE_CELL_OBSERVATION")
	}
	if observed == nil {
		return unknownResult(result, "LOCATE_EVIDENCE", "CELL_NOT_OBSERVED")
	}
	result.Producer, result.Step, result.Reason = observed.Producer, observed.Step, observed.Reason
	result.EvidenceKey, result.EvidencePath, result.EvidenceDigest = observed.EvidenceKey, observed.EvidencePath, observed.EvidenceDigest
	switch observed.State {
	case StateClosed:
		if !validReference(*observed) {
			return unknownResult(result, "VALIDATE_EVIDENCE", "EVIDENCE_REFERENCE_INVALID")
		}
		if useCase.ID == "debugging" && (stage.ID == "DETERMINISTIC_REPLAY" || stage.ID == "RESOURCE_OBSERVED") {
			if reason := validateDebugBinding(*observed, graph, stage.ID); reason != "" {
				return refutedResult(result, "VERIFY_GOOO_GRAPH", reason)
			}
		}
		result.State, result.ClaimStatus, result.Resolution = StateClosed, "DISCHARGED", "EXACT"
	case StateOpen:
		if observed.Producer == "" || observed.Step == "" || observed.Reason == "" {
			return unknownResult(result, "CLASSIFY_GAP", "OPEN_COORDINATE_INCOMPLETE")
		}
		result.State, result.ClaimStatus, result.Resolution = StateOpen, "OPEN", "EXACT"
	case StateRefuted:
		if !validReference(*observed) || observed.Reason == "" {
			return unknownResult(result, "VALIDATE_REFUTATION", "REFUTATION_REFERENCE_INVALID")
		}
		result.State, result.ClaimStatus, result.Resolution = StateRefuted, "REFUTED", "EXACT"
	case StateUnknown:
		return unknownResult(result, fallback(observed.Step, "CLASSIFY_EVIDENCE"), fallback(observed.Reason, "EVIDENCE_STATE_UNKNOWN"))
	default:
		return unknownResult(result, "CLASSIFY_EVIDENCE", "EVIDENCE_STATE_UNKNOWN")
	}
	return result
}

func unknownResult(result CellResult, step, reason string) CellResult {
	result.State, result.ClaimStatus, result.Resolution = StateUnknown, "OPEN", "LOWER_RESOLUTION"
	result.Step, result.Reason = step, reason
	return result
}

func validReference(value CellObservation) bool {
	if value.Producer == "" || value.Step == "" || value.Reason == "" || value.EvidenceKey == "" || value.EvidencePath == "" || len(value.EvidenceDigest) != 71 || !strings.HasPrefix(value.EvidenceDigest, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value.EvidenceDigest, "sha256:"))
	return err == nil
}

func fallback(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
