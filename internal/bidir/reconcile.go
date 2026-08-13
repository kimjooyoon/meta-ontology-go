package bidir

import "fmt"

// ProjectFacts projects semantic relations into deterministic facts.
func ProjectFacts(model Model) FactSet {
	kinds := make(map[ID]Kind, len(model.Nodes))
	for _, node := range model.Nodes {
		kinds[node.ID] = node.Kind
	}
	var facts FactSet
	for _, relation := range model.Relations {
		facts = append(facts, Fact{Layer: DeterministicFact, Subject: relation.Source, Predicate: relation.Kind, Object: relation.Target, SubjectKind: kinds[relation.Source], ObjectKind: kinds[relation.Target], Attributes: cloneStringMap(relation.Attributes), Source: relation.Span})
	}
	return facts.Normalized()
}

// LiftFacts is the strict Go-fact entry point.
func LiftFacts(base Model, facts FactSet) (ReconcileResult, error) {
	return Reconcile(base, FactDelta{Added: facts})
}

// Reconcile applies a source-view fact delta under strict policy.
func Reconcile(base Model, changes FactDelta) (ReconcileResult, error) {
	return ReconcileWithOptions(base, changes, DefaultReconcileOptions())
}

// ReconcileWithOptions applies a fact delta transactionally.
func ReconcileWithOptions(base Model, changes FactDelta, options ReconcileOptions) (ReconcileResult, error) {
	rawObservation := newRawFactObservation(changes)
	base = base.Normalized()
	if err := base.Validate(); err != nil {
		return ReconcileResult{Model: base, RawObservation: rawObservation}, err
	}
	result := ReconcileResult{Model: base, RawObservation: rawObservation}
	for _, fact := range append(append(FactSet(nil), changes.Removed...), changes.Added...) {
		if conflict := validateFactInput(fact); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
		}
	}
	if len(result.Conflicts) > 0 {
		return result, &ReconcileError{Conflicts: result.Conflicts}
	}
	working := base.Clone()
	changes = changes.Normalized()
	for _, fact := range changes.Removed {
		if conflict := removeFact(&working, fact, options); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
			continue
		}
		result.Accepted = append(result.Accepted, fact)
	}
	for _, fact := range changes.Added {
		if conflict := addFact(&working, &result, fact, options); conflict != nil {
			result.Conflicts = append(result.Conflicts, *conflict)
		}
	}
	if len(result.Conflicts) > 0 {
		result.Model = base
		return result, &ReconcileError{Conflicts: result.Conflicts}
	}
	working.Candidates.Normalize()
	working.Normalize()
	result.Model = working
	result.Accepted.Normalize()
	result.Syntactic.Normalize()
	result.Candidates = working.Candidates.ByLayer(CandidateFact)
	result.Delta = Diff(base, working)
	result.Locality = LocalityForDelta(base, result.Delta)
	return result, nil
}

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
		// Keep the observation in the result/evidence stream, but mirror the
		// semantic graph rule that an existing deterministic fact shadows a
		// candidate with the same triple.
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

func findRelation(model Model, predicate Predicate, source, target ID) (Relation, bool) {
	for _, relation := range model.Relations {
		if relation.Kind == predicate && relation.Source == source && relation.Target == target {
			return relation, true
		}
	}
	return Relation{}, false
}

func removeRelation(relations []Relation, predicate Predicate, source, target ID) []Relation {
	result := relations[:0]
	for _, relation := range relations {
		if relation.Kind == predicate && relation.Source == source && relation.Target == target {
			continue
		}
		result = append(result, relation)
	}
	return result
}

func ensureEndpoint(model *Model, id ID, hintedKind Kind, fact Fact, subject bool) *Conflict {
	for _, node := range model.Nodes {
		if node.ID != id {
			continue
		}
		if hintedKind != "" && node.Kind != hintedKind {
			return &Conflict{Kind: ConflictKindMismatch, Fact: fact, Message: fmt.Sprintf("%s %q is %s, fact says %s", endpointLabel(subject), id, node.Kind, hintedKind)}
		}
		return nil
	}
	return &Conflict{Kind: ConflictUnknownEndpoint, Fact: fact, Message: fmt.Sprintf("%s %q is not registered in the base model", endpointLabel(subject), id)}
}

func endpointLabel(subject bool) string {
	if subject {
		return "subject"
	}
	return "object"
}
