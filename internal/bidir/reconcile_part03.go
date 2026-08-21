package bidir

import (
	"fmt"
)

func addDeterministicFact(model *Model, fact Fact, options ReconcileOptions) *Conflict {
	if !knownSemanticPredicate(fact.Predicate) {
		return &Conflict{Kind: ConflictUnknownPredicate, Fact: fact, Message: fmt.Sprintf("predicate %q is not deterministic semantic vocabulary", fact.Predicate)}
	}
	if err := validateID(fact.Subject); err != nil {
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("invalid subject: %v", err)}
	}
	if err := validateID(fact.Object); err != nil {
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("invalid object: %v", err)}
	}
	if existing, exists := findRelation(*model, fact.Predicate, fact.Subject, fact.Object); exists {
		if relationSemanticEqual(existing, relationFromFact(fact)) {
			return nil
		}
		if options.RequireSource && !fact.Source.Valid() {
			return &Conflict{Kind: ConflictMissingSource, Fact: fact, Message: "changing relation attributes requires source-backed evidence"}
		}
		model.Relations = removeRelation(model.Relations, fact.Predicate, fact.Subject, fact.Object)
	}
	if options.RequireSource && !fact.Source.Valid() {
		return &Conflict{Kind: ConflictMissingSource, Fact: fact, Message: "new semantic relation requires source-backed evidence"}
	}
	if conflict := ensureEndpoint(model, fact.Subject, fact.SubjectKind, fact, true); conflict != nil {
		return conflict
	}
	if conflict := ensureEndpoint(model, fact.Object, fact.ObjectKind, fact, false); conflict != nil {
		return conflict
	}
	model.Relations = append(model.Relations, relationFromFact(fact))
	return nil
}
func removeFact(model *Model, fact Fact, options ReconcileOptions) *Conflict {
	switch fact.Layer {
	case SyntacticFact:
		return nil
	case CandidateFact:
		model.Candidates = model.Candidates.withoutKey(fact.Key())
		return nil
	case DeterministicFact:
		if !knownSemanticPredicate(fact.Predicate) {
			return &Conflict{Kind: ConflictUnknownPredicate, Fact: fact, Message: fmt.Sprintf("predicate %q is not deterministic semantic vocabulary", fact.Predicate)}
		}
		if _, exists := findRelation(*model, fact.Predicate, fact.Subject, fact.Object); !exists {
			return nil
		}
		if options.RequireSource && !fact.Source.Valid() {
			return &Conflict{Kind: ConflictMissingSource, Fact: fact, Message: "removing a semantic relation requires source-backed evidence"}
		}
		model.Relations = removeRelation(model.Relations, fact.Predicate, fact.Subject, fact.Object)
		return nil
	default:
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("unsupported fact layer %d", fact.Layer)}
	}
}
func knownSemanticPredicate(predicate Predicate) bool {
	switch predicate {
	case PredicateUsed, PredicateWasGeneratedBy, PredicateWasDerivedFrom, PredicateInvokes, PredicateRepresents, PredicateSpecialization:
		return true
	default:
		return false
	}
}
func relationFromFact(fact Fact) Relation {
	return Relation{ID: StableRelationID(fact.Predicate, fact.Subject, fact.Object), Kind: fact.Predicate, Source: fact.Subject, Target: fact.Object, Attributes: cloneStringMap(fact.Attributes), Span: fact.Source}
}
