package bidir

import (
	"fmt"
)

func validateFactInput(fact Fact) *Conflict {
	if err := fact.Source.Validate(); err != nil {
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("invalid source span: %v", err)}
	}
	if fact.Layer != DeterministicFact {
		if fact.Layer != SyntacticFact && fact.Layer != CandidateFact {
			return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("unsupported fact layer %d", fact.Layer)}
		}
		return nil
	}
	if !knownSemanticPredicate(fact.Predicate) {
		return &Conflict{Kind: ConflictUnknownPredicate, Fact: fact, Message: fmt.Sprintf("predicate %q is not deterministic semantic vocabulary", fact.Predicate)}
	}
	if err := validateID(fact.Subject); err != nil {
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("invalid subject: %v", err)}
	}
	if err := validateID(fact.Object); err != nil {
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("invalid object: %v", err)}
	}
	return nil
}
func addFact(model *Model, result *ReconcileResult, fact Fact, options ReconcileOptions) *Conflict {
	switch fact.Layer {
	case SyntacticFact:
		result.Syntactic = append(result.Syntactic, fact)
		return nil
	case CandidateFact:
		candidate := fact.normalized()

		if _, exists := findRelation(*model, fact.Predicate, fact.Subject, fact.Object); !exists {
			model.Candidates = append(model.Candidates, candidate)
		}
		result.Candidates = append(result.Candidates, fact)
		return nil
	case DeterministicFact:
		if conflict := addDeterministicFact(model, fact, options); conflict != nil {
			return conflict
		}
		model.Candidates = model.Candidates.withoutSemanticKey(fact.SemanticKey())
		result.Accepted = append(result.Accepted, fact)
		return nil
	default:
		return &Conflict{Kind: ConflictInvalidFact, Fact: fact, Message: fmt.Sprintf("unsupported fact layer %d", fact.Layer)}
	}
}
