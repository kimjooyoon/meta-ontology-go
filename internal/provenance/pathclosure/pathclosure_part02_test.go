package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func requirementForPath(pathID string, edges []semantic.InferenceEdge) pathclosure.Requirement {
	records := make([]semantic.ID, 0, len(edges))
	kinds := make([]semantic.InferenceKind, 0, len(edges))
	for _, edge := range edges {
		records = append(records, edge.RecordID)
		kinds = append(kinds, edge.Kind)
	}
	return pathclosure.Requirement{
		PathID:        fixtureID("path/" + pathID),
		RecordIDs:     records,
		ExpectedKinds: kinds,
		StartID:       edges[0].SubjectID,
		EndID:         edges[len(edges)-1].ObjectID,
	}
}
func assertEvaluation(t *testing.T, got pathclosure.Result, status pathclosure.Status, code string, numerator, denominator int) {
	t.Helper()
	if got.Status != status || got.Code != code {
		t.Fatalf("evaluation status/code = %s/%s, want %s/%s: %#v", got.Status, got.Code, status, code, got)
	}
	if got.Numerator != numerator || got.Denominator != denominator {
		t.Fatalf("evaluation coverage = %d/%d, want %d/%d: %#v", got.Numerator, got.Denominator, numerator, denominator, got)
	}
}
