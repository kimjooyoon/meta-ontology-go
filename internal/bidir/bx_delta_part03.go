package bidir

func factCanonicalValues(facts FactSet) []string {
	values := make([]string, len(facts))
	for index, fact := range facts {
		values[index] = factCanonical(fact)
	}
	return values
}
func candidateFacts(facts FactSet) FactSet {
	candidates := make(FactSet, 0)
	for _, fact := range facts {
		if fact.Layer == CandidateFact {
			candidates = append(candidates, fact)
		}
	}
	return candidates
}
func spanTexts(spans []SourceSpan) []string {
	texts := make([]string, len(spans))
	for index, span := range spans {
		texts[index] = spanText(span)
	}
	return texts
}
func sameLocality(left, right Locality) bool {
	return sameIDs(left.Touched, right.Touched) && sameIDs(left.Affected, right.Affected)
}
func detachedLocality(locality Locality) Locality {
	return Locality{Touched: append([]ID(nil), locality.Touched...), Affected: append([]ID(nil), locality.Affected...)}
}
