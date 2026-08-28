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
	result.EvidenceKey, result.EvidencePath = observed.EvidenceKey, observed.EvidencePath
	result.EvidenceDigest = observed.EvidenceDigest
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
		return unknownResult(result, fallback(observed.Step, "CLASSIFY_EVIDENCE"),
			fallback(observed.Reason, "EVIDENCE_STATE_UNKNOWN"))
	default:
		return unknownResult(result, "CLASSIFY_EVIDENCE", "EVIDENCE_STATE_UNKNOWN")
	}
	return result
}

func refutedResult(result CellResult, step, reason string) CellResult {
	result.State, result.ClaimStatus, result.Resolution = StateRefuted, "REFUTED", "EXACT"
	result.Step, result.Reason = step, reason
	return result
}

func validateDebugBinding(observed CellObservation, graph GraphObservation, stage string) string {
	activityID := "languageutility://activity/observe-debugging-" + strings.ToLower(strings.ReplaceAll(stage, "_", "-"))
	inputID := "gooo://meta/language-utility/entity/cell"
	outputID := "gooo://meta/language-utility/entity/evidence"
	if graph.Schema != "gooo-graph/v1" || graph.ActivityCount != 44 || graph.EdgeCount != 88 ||
		graph.DebugActivityCount != 2 || graph.DebugOutputCount != 2 || graph.DebugUsedEdgeCount != 2 ||
		graph.DebugGeneratedEdgeCount != 2 {
		return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
	}
	expectedActivities := map[string]bool{
		"languageutility://activity/observe-debugging-deterministic-replay": true,
		"languageutility://activity/observe-debugging-resource-observed":    true,
	}
	if len(graph.DebugActivityIDs) != len(expectedActivities) {
		return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
	}
	seenActivities := map[string]bool{}
	for _, activity := range graph.DebugActivityIDs {
		if !expectedActivities[activity] || seenActivities[activity] {
			return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
		}
		seenActivities[activity] = true
	}
	expectedEdges := map[string]bool{
		"used\x00languageutility://activity/observe-debugging-deterministic-replay\x00gooo://meta/language-utility/entity/cell":               true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-deterministic-replay": true,
		"used\x00languageutility://activity/observe-debugging-resource-observed\x00gooo://meta/language-utility/entity/cell":                  true,
		"wasGeneratedBy\x00gooo://meta/language-utility/entity/evidence\x00languageutility://activity/observe-debugging-resource-observed":    true,
	}
	if len(graph.DebugCausalEdges) != len(expectedEdges) {
		return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
	}
	seenGraphEdges := map[string]bool{}
	for _, edge := range graph.DebugCausalEdges {
		key := edge.Relation + "\x00" + edge.Subject + "\x00" + edge.Object
		if !expectedEdges[key] || seenGraphEdges[key] {
			return "GOOO_GRAPH_ACTIVITY_OR_EDGE_CONTRADICTION"
		}
		seenGraphEdges[key] = true
	}
	if observed.MetaActivityID != activityID || observed.MetaInputID != inputID || observed.MetaOutputID != outputID ||
		observed.ActivityMatches != 1 || observed.OutputMatches != 1 || observed.UsedEdgeMatches != 1 ||
		observed.GeneratedEdgeMatches != 1 || len(observed.CausalEdges) != 2 {
		return "GOOO_META_ACTIVITY_OUTPUT_OR_EDGE_MISSING"
	}
	seen := map[string]bool{}
	for _, edge := range observed.CausalEdges {
		key := edge.Relation + "\x00" + edge.Subject + "\x00" + edge.Object
		if seen[key] {
			return "GOOO_CAUSAL_EDGE_DUPLICATE"
		}
		seen[key] = true
	}
	if !seen["used\x00"+activityID+"\x00"+inputID] ||
		!seen["wasGeneratedBy\x00"+outputID+"\x00"+activityID] {
		return "GOOO_CAUSAL_EDGE_MISSING"
	}
	return ""
}

func unknownResult(result CellResult, step, reason string) CellResult {
	result.State, result.ClaimStatus, result.Resolution = StateUnknown, "OPEN", "LOWER_RESOLUTION"
	result.Step, result.Reason = step, reason
	return result
}

func validReference(value CellObservation) bool {
	if value.Producer == "" || value.Step == "" || value.Reason == "" ||
		value.EvidenceKey == "" || value.EvidencePath == "" || len(value.EvidenceDigest) != 71 ||
		!strings.HasPrefix(value.EvidenceDigest, "sha256:") {
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
