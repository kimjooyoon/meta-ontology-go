package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func prepareRequirements(requirements []Requirement) ([]requirementState, []semantic.ID) {
	states := make([]requirementState, 0, len(requirements))
	required := make([]semantic.ID, 0, len(requirements))
	seen := make(map[semantic.ID]struct{}, len(requirements))
	for _, raw := range requirements {
		state := requirementState{raw: raw, normalized: raw}
		if pathID, err := semantic.ParseIdentity(raw.PathID.String()); err == nil {
			state.normalized.PathID = pathID
		}
		normalized, err := normalizeRequirement(raw)
		if err != nil {
			state.malformed = true
		} else {
			state.normalized = normalized
		}
		if hasDuplicateRecordIDs(raw.RecordIDs) {
			state.duplicate = true
		}
		if state.normalized.PathID != "" {
			if _, exists := seen[state.normalized.PathID]; exists {
				state.duplicate = true
			} else {
				seen[state.normalized.PathID] = struct{}{}
			}
		}
		required = appendID(required, state.normalized.PathID)
		states = append(states, state)
	}
	sortIDs(required)
	return states, required
}
func hasDuplicateRecordIDs(ids []semantic.ID) bool {
	seen := make(map[semantic.ID]struct{}, len(ids))
	for _, rawID := range ids {
		id := rawID
		if normalized, err := semantic.ParseIdentity(rawID.String()); err == nil {
			id = normalized
		}
		if _, exists := seen[id]; exists {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
