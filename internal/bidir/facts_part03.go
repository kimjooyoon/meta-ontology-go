package bidir

// withoutSemanticKey removes only candidate observations for an authoritative
// triple. Candidate identity includes its layer, but shadowing is deliberately
// layer-independent: a deterministic fact must not leave a stale candidate in
// the model while the raw/evidence boundary retains the observation.
func (s FactSet) withoutSemanticKey(key string) FactSet {
	result := make(FactSet, 0, len(s))
	for _, fact := range s {
		if fact.Layer == CandidateFact && fact.SemanticKey() == key {
			continue
		}
		result = append(result, fact)
	}
	return result
}
func factLess(left, right Fact) bool {
	if left.Layer != right.Layer {
		return left.Layer < right.Layer
	}
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	if left.Object != right.Object {
		return left.Object < right.Object
	}
	if left.Source.File != right.Source.File {
		return left.Source.File < right.Source.File
	}
	if left.Source.Start != right.Source.Start {
		return left.Source.Start < right.Source.Start
	}
	if left.Source.End != right.Source.End {
		return left.Source.End < right.Source.End
	}
	if left.Source.StartLine != right.Source.StartLine {
		return left.Source.StartLine < right.Source.StartLine
	}
	if left.Source.StartColumn != right.Source.StartColumn {
		return left.Source.StartColumn < right.Source.StartColumn
	}
	if left.Source.EndLine != right.Source.EndLine {
		return left.Source.EndLine < right.Source.EndLine
	}
	if left.Source.EndColumn != right.Source.EndColumn {
		return left.Source.EndColumn < right.Source.EndColumn
	}
	if left.SubjectKind != right.SubjectKind {
		return left.SubjectKind < right.SubjectKind
	}
	if left.ObjectKind != right.ObjectKind {
		return left.ObjectKind < right.ObjectKind
	}
	if left.Reason != right.Reason {
		return left.Reason < right.Reason
	}
	return attributesLess(left.Attributes, right.Attributes)
}
func attributesLess(left, right map[string]string) bool {
	leftKeys, rightKeys := mapKeys(left), mapKeys(right)
	for index := 0; index < len(leftKeys) && index < len(rightKeys); index++ {
		if leftKeys[index] != rightKeys[index] {
			return leftKeys[index] < rightKeys[index]
		}
		if left[leftKeys[index]] != right[rightKeys[index]] {
			return left[leftKeys[index]] < right[rightKeys[index]]
		}
	}
	return len(leftKeys) < len(rightKeys)
}
