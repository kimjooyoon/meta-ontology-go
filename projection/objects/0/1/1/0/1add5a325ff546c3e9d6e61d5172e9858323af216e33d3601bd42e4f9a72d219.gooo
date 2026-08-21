package bidir

import (
	"sort"
)

func (f Fact) normalized() Fact {
	f.Attributes = cloneStringMap(f.Attributes)
	return f
}

// FactSet is a deterministic set-like collection.
type FactSet []Fact

// Normalized returns a sorted, deduplicated copy.
func (s FactSet) Normalized() FactSet {
	copySet := make(FactSet, len(s))
	for index, fact := range s {
		copySet[index] = fact.normalized()
	}
	sort.SliceStable(copySet, func(i, j int) bool { return factLess(copySet[i], copySet[j]) })
	result := make(FactSet, 0, len(copySet))
	seen := make(map[FactKey]struct{}, len(copySet))
	for _, fact := range copySet {
		if _, exists := seen[fact.Key()]; !exists {
			seen[fact.Key()] = struct{}{}
			result = append(result, fact)
		}
	}
	return result
}

// Normalize sorts and deduplicates the set in place.
func (s *FactSet) Normalize() {
	if s != nil {
		*s = s.Normalized()
	}
}

// ByLayer returns the deterministic subset with the requested layer.
func (s FactSet) ByLayer(layer FactLayer) FactSet {
	var result FactSet
	for _, fact := range s {
		if fact.Layer == layer {
			result = append(result, fact)
		}
	}
	return result.Normalized()
}

// Contains reports whether a fact with the same key exists.
func (s FactSet) Contains(candidate Fact) bool {
	key := candidate.Key()
	for _, fact := range s {
		if fact.Key() == key {
			return true
		}
	}
	return false
}
func (s FactSet) withoutKey(key FactKey) FactSet {
	result := make(FactSet, 0, len(s))
	for _, fact := range s {
		if fact.Key() != key {
			result = append(result, fact)
		}
	}
	return result
}
