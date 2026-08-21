package query

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
