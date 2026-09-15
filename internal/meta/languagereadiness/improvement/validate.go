package improvement

import "strings"

func inspect(snapshot Snapshot) inspection {
	state := inspection{statuses: make(map[string]EvidenceStatus)}
	switch {
	case snapshot.ContractSchema != SnapshotSchema:
		state.reason = "CONTRACT_SCHEMA_UNKNOWN"
	case !validDigest(snapshot.RegistryDigest):
		state.reason = "REGISTRY_DIGEST_INVALID"
	case snapshot.Total != SnapshotTotal || int64(len(snapshot.Evidence)) != SnapshotTotal:
		state.reason = "DENOMINATOR_INVALID"
	}
	if state.reason != "" {
		return state
	}
	var completed int64
	for _, evidence := range snapshot.Evidence {
		if evidence.ID == "" || evidence.ID != strings.TrimSpace(evidence.ID) {
			state.reason = "EVIDENCE_ID_INVALID"
			return state
		}
		if _, exists := state.statuses[evidence.ID]; exists {
			state.reason = "EVIDENCE_ID_DUPLICATE"
			return state
		}
		switch evidence.Status {
		case Satisfied:
			completed++
		case NotSatisfied:
		case Unresolved:
			state.unresolved++
		default:
			state.reason = "EVIDENCE_STATUS_UNKNOWN"
			return state
		}
		state.statuses[evidence.ID] = evidence.Status
	}
	switch {
	case snapshot.Completed != completed:
		state.reason = "COMPLETED_COUNT_INVALID"
	case snapshot.BasisPoints != completed*10_000/SnapshotTotal:
		state.reason = "BASIS_POINTS_INVALID"
	}
	return state
}
