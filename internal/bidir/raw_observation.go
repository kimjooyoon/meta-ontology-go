package bidir

import (
	"sort"
	"strings"
)

func newRawFactObservation(delta FactDelta) RawFactObservation {
	observation := RawFactObservation{
		Added:   rawFacts(delta.Added),
		Removed: rawFacts(delta.Removed),
	}
	observation.EvidenceHash = rawObservationHash(observation)
	return observation
}

func rawFacts(facts FactSet) FactSet {
	copyFacts := make(FactSet, len(facts))
	for index, fact := range facts {
		copyFacts[index] = fact.normalized()
	}
	sort.SliceStable(copyFacts, func(i, j int) bool {
		return rawFactLess(copyFacts[i], copyFacts[j])
	})
	return copyFacts
}

func rawFactLess(left, right Fact) bool {
	if factLess(left, right) {
		return true
	}
	if factLess(right, left) {
		return false
	}
	return left.EvidenceID < right.EvidenceID
}

func rawObservationHash(observation RawFactObservation) string {
	var builder strings.Builder
	writeRawFacts(&builder, "added", observation.Added)
	writeRawFacts(&builder, "removed", observation.Removed)
	return digest(builder.String())
}

func writeRawFacts(builder *strings.Builder, label string, facts FactSet) {
	writePart(builder, label)
	for _, fact := range facts {
		writePart(builder, fact.EvidenceID)
		writePart(builder, factCanonical(fact))
	}
}
