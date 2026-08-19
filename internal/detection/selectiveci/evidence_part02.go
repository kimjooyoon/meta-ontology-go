package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func evaluatePath(path ProvenancePath) error {
	requirement, err := pathRequirement(path.Requirement)
	if err != nil {
		return failure(ReasonAmbiguousPath, err.Error())
	}
	if normalized, normalizeErr := path.Path.Normalized(); normalizeErr == nil {
		if topologyErr := validatePathTopology(normalized, requirement); topologyErr != nil {
			return topologyErr
		}
	}
	evaluation := pathclosure.Evaluate(path.Path, []pathclosure.Requirement{requirement})
	if evaluation.Status == pathclosure.PASS && len(evaluation.Complete) == 1 {
		return nil
	}
	if evaluation.Code == pathclosure.CodeDuplicate {
		return failure(ReasonDuplicateID, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingRecord {
		return failure(ReasonUnknownPath, evaluation.Code)
	}
	if evaluation.Code == pathclosure.CodeMissingEvidence || evaluation.Code == pathclosure.CodeMissingSnapshot {
		return failure(ReasonDanglingReference, evaluation.Code)
	}
	if strings.Contains(strings.ToLower(evaluation.Code), "malformed") {
		return failure(ReasonAmbiguousPath, evaluation.Code)
	}
	return failure(ReasonEvaluatorError, evaluation.Code)
}
func validatePathTopology(path semantic.InferencePathV1, requirement pathclosure.Requirement) error {
	byID := make(map[semantic.ID]semantic.InferenceEdge, len(path.Edges))
	for _, edge := range path.Edges {
		byID[edge.RecordID] = edge
	}
	edges := make([]semantic.InferenceEdge, 0, len(requirement.RecordIDs))
	for _, recordID := range requirement.RecordIDs {
		edge, ok := byID[recordID]
		if !ok {
			return nil
		}
		edges = append(edges, edge)
	}
	if _, err := semantic.NewInferencePathChain(edges...); err != nil {
		message := strings.ToLower(err.Error())
		if strings.Contains(message, "cycle") || strings.Contains(message, "path_orphan") {
			return failure(ReasonCycle, err.Error())
		}
		if strings.Contains(message, "path_ambiguity") {
			return failure(ReasonAmbiguousPath, err.Error())
		}
	}
	return nil
}
