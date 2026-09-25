package bidir

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func semanticPredicate(predicate Predicate) (semantic.Relation, bool) {
	switch predicate {
	case PredicateUsed:
		return semantic.Used, true
	case PredicateWasGeneratedBy:
		return semantic.WasGeneratedBy, true
	case PredicateWasDerivedFrom:
		return semantic.WasDerivedFrom, true
	default:
		return "", false
	}
}
