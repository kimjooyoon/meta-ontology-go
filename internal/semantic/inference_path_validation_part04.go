package semantic

import (
	"fmt"
)

func NewInferencePathChain(edges ...InferenceEdge) (InferencePathChain, error) {
	if len(edges) == 0 {
		return InferencePathChain{}, fmt.Errorf("%w: path_orphan: empty path", ErrInferencePath)
	}
	normalized := make([]InferenceEdge, 0, len(edges))
	bySubject := make(map[ID][]InferenceEdge, len(edges))
	objects := make(map[ID]struct{}, len(edges))
	seen := make(map[ID]struct{}, len(edges))
	for _, raw := range edges {
		edge, err := raw.normalized()
		if err != nil {
			return InferencePathChain{}, fmt.Errorf("%w: path_orphan: %v", ErrInferencePath, err)
		}
		if _, exists := seen[edge.RecordID]; exists {
			return InferencePathChain{}, fmt.Errorf("%w: path_ambiguity: duplicate edge %s", ErrInferencePath, edge.RecordID)
		}
		seen[edge.RecordID] = struct{}{}
		normalized = append(normalized, edge)
		bySubject[edge.SubjectID] = append(bySubject[edge.SubjectID], edge)
		objects[edge.ObjectID] = struct{}{}
	}
	starts := make([]ID, 0, len(bySubject))
	for subject := range bySubject {
		if _, hasIncoming := objects[subject]; !hasIncoming {
			starts = append(starts, subject)
		}
	}
	if len(starts) != 1 {
		return InferencePathChain{}, fmt.Errorf("%w: path_ambiguity: want one start, got %d", ErrInferencePath, len(starts))
	}
	ordered := make([]InferenceEdge, 0, len(edges))
	visited := make(map[ID]struct{}, len(edges))
	current := starts[0]
	for {
		outgoing := bySubject[current]
		if len(outgoing) == 0 {
			break
		}
		if len(outgoing) != 1 {
			return InferencePathChain{}, fmt.Errorf(
				"%w: path_ambiguity: %s has %d outgoing edges", ErrInferencePath, current, len(outgoing),
			)
		}
		edge := outgoing[0]
		if _, exists := visited[edge.RecordID]; exists {
			return InferencePathChain{}, fmt.Errorf("%w: path_orphan: cycle at %s", ErrInferencePath, edge.RecordID)
		}
		visited[edge.RecordID] = struct{}{}
		ordered = append(ordered, edge)
		current = edge.ObjectID
	}
	if len(ordered) != len(normalized) {
		return InferencePathChain{}, fmt.Errorf(
			"%w: path_orphan: %d edges are disconnected", ErrInferencePath, len(normalized)-len(ordered),
		)
	}
	return InferencePathChain{Edges: ordered}, nil
}
func (c InferencePathChain) Validate() error {
	_, err := NewInferencePathChain(c.Edges...)
	return err
}
