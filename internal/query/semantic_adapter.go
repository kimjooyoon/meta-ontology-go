package query

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// FromSemanticIR builds a read-only query projection from a validated semantic
// IR. The semantic package remains authoritative; Graph is only a detached
// query view. Invalid IR, including unknown endpoints or invalid relation
// kinds, is rejected before any projection fact is added.
func FromSemanticIR(ir semantic.IR) (*Graph, error) {
	if err := ir.Validate(); err != nil {
		return nil, fmt.Errorf("semantic IR is not queryable: %w", err)
	}
	normalized, err := ir.Normalized()
	if err != nil {
		return nil, fmt.Errorf("semantic IR cannot be normalized for query: %w", err)
	}
	graph := New()
	evidenceStatus := "known_empty"
	if len(normalized.Evidence()) > 0 {
		evidenceStatus = "available"
	}
	provenanceStatus := "known_empty"
	if len(normalized.Evidence()) > 0 {
		provenanceStatus = "available"
	}
	graph.binding = &projectionBinding{
		semanticDigest:   normalized.StableHash(),
		sourceStatus:     "unavailable",
		evidenceDigest:   normalized.EvidenceHash(),
		provenanceDigest: normalized.ProvenanceHash(),
		evidenceStatus:   evidenceStatus,
		provenanceStatus: provenanceStatus,
	}
	for _, fact := range normalized.Graph.AllFacts() {
		projected, err := projectSemanticFact(fact)
		if err != nil {
			return nil, err
		}
		if err := graph.Add(projected); err != nil {
			return nil, fmt.Errorf("project semantic fact: %w", err)
		}
	}
	return graph, nil
}

func projectSemanticFact(fact semantic.Fact) (Fact, error) {
	predicate, err := ParseRelation(Relation(fact.Predicate))
	if err != nil {
		return Fact{}, fmt.Errorf("semantic relation %q is not queryable: %w", fact.Predicate, err)
	}
	status := FactDeterministic
	if fact.Status == semantic.FactCandidate {
		status = FactCandidate
	}
	if fact.Status != semantic.FactDeterministic && fact.Status != semantic.FactCandidate {
		return Fact{}, fmt.Errorf("semantic fact has unsupported status %d", fact.Status)
	}
	return Fact{
		Subject:   ID(fact.Subject.String()),
		Predicate: predicate,
		Object:    ID(fact.Object.String()),
		Status:    status,
		Reason:    fact.Reason,
	}, nil
}
