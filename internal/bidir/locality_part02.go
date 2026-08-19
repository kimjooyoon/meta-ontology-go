package bidir

import (
	"sort"
)

// LocalityForDelta computes changed endpoints and their old one-hop neighbors.
func LocalityForDelta(base Model, delta Delta) Locality {
	touched := make(map[ID]struct{})
	for _, node := range append(delta.AddedNodes, delta.RemovedNodes...) {
		touched[node.ID] = struct{}{}
	}
	for _, relation := range append(delta.AddedRelations, delta.RemovedRelations...) {
		touched[relation.Source] = struct{}{}
		touched[relation.Target] = struct{}{}
	}
	affected := make(map[ID]struct{}, len(touched))
	for id := range touched {
		affected[id] = struct{}{}
	}
	for _, relation := range base.Relations {
		if _, exists := touched[relation.Source]; exists {
			affected[relation.Target] = struct{}{}
		}
		if _, exists := touched[relation.Target]; exists {
			affected[relation.Source] = struct{}{}
		}
	}
	return Locality{Touched: sortedIDs(touched), Affected: sortedIDs(affected)}
}

// LocalityBetween computes the region changed between two models.
func LocalityBetween(before, after Model) Locality {
	return LocalityForDelta(before, Diff(before, after))
}

// Contains reports whether an ID is touched or affected.
func (l Locality) Contains(id ID) bool {
	for _, candidate := range l.Affected {
		if candidate == id {
			return true
		}
	}
	return false
}
func sortedIDs(values map[ID]struct{}) []ID {
	ids := make([]ID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
