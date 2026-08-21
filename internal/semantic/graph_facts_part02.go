package semantic

func mergeFact(existing, incoming Fact) Fact {
	merged := existing
	if merged.Span.IsZero() || (!incoming.Span.IsZero() && spanLess(incoming.Span, merged.Span)) {
		merged.Span = incoming.Span
	}
	if merged.Reason == "" || (incoming.Reason != "" && incoming.Reason < merged.Reason) {
		merged.Reason = incoming.Reason
	}
	return merged
}
func spanLess(left, right Span) bool {
	if left.File != right.File {
		return left.File < right.File
	}
	if left.Start.Offset != right.Start.Offset {
		return left.Start.Offset < right.Start.Offset
	}
	if left.Start.Line != right.Start.Line {
		return left.Start.Line < right.Start.Line
	}
	if left.Start.Column != right.Start.Column {
		return left.Start.Column < right.Start.Column
	}
	if left.End.Offset != right.End.Offset {
		return left.End.Offset < right.End.Offset
	}
	if left.End.Line != right.End.Line {
		return left.End.Line < right.End.Line
	}
	return left.End.Column < right.End.Column
}
func (g Graph) Facts() []Fact {
	return g.DeterministicFacts()
}
func (g Graph) DeterministicFacts() []Fact {
	facts := make([]Fact, 0, len(g.facts))
	for _, fact := range g.facts {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}
func (g Graph) Candidates() []Fact {
	facts := make([]Fact, 0, len(g.candidates))
	for _, fact := range g.candidates {
		facts = append(facts, fact)
	}
	sortFacts(facts)
	return facts
}

// SortedFacts is an explicit alias for adapters that prefer sorted snapshots.
func (g Graph) SortedFacts() []Fact {
	return g.AllFacts()
}
