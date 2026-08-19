package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// selectedOracleChain independently orders the exact edge IDs named by a
// receipt. It intentionally does not call semantic path normalization: the
// oracle owns the closure, start/end, fork, cycle, and disconnected checks.
func selectedOracleChain(ids []string, edges map[semantic.ID]semantic.InferenceEdge) ([]semantic.InferenceEdge, bool) {
	if len(ids) == 0 {
		return nil, false
	}
	selected := make(map[semantic.ID]semantic.InferenceEdge, len(ids))
	for _, rawID := range ids {
		id, err := semantic.ParseIdentity(rawID)
		if err != nil {
			return nil, false
		}
		if _, duplicate := selected[id]; duplicate {
			return nil, false
		}
		edge, ok := edges[id]
		if !ok {
			return nil, false
		}
		selected[id] = edge
	}
	bySubject := make(map[semantic.ID][]semantic.InferenceEdge, len(selected))
	objects := make(map[semantic.ID]struct{}, len(selected))
	for _, edge := range selected {
		bySubject[edge.SubjectID] = append(bySubject[edge.SubjectID], edge)
		objects[edge.ObjectID] = struct{}{}
	}
	var start semantic.ID
	starts := 0
	for subject := range bySubject {
		if _, hasIncoming := objects[subject]; !hasIncoming {
			start = subject
			starts++
		}
	}
	if starts != 1 {
		return nil, false
	}
	ordered := make([]semantic.InferenceEdge, 0, len(selected))
	visited := make(map[semantic.ID]struct{}, len(selected))
	for {
		outgoing := bySubject[start]
		if len(outgoing) == 0 {
			break
		}
		if len(outgoing) != 1 {
			return nil, false
		}
		edge := outgoing[0]
		if _, duplicate := visited[edge.RecordID]; duplicate {
			return nil, false
		}
		visited[edge.RecordID] = struct{}{}
		ordered = append(ordered, edge)
		start = edge.ObjectID
	}
	if len(ordered) != len(selected) {
		return nil, false
	}
	return ordered, true
}
