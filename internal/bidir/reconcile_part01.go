package bidir

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
