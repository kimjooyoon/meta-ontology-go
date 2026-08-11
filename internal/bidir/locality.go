package bidir

import "sort"

// Apply applies a delta transactionally and validates the resulting model.
func (m Model) Apply(delta Delta) (Model, error) {
	working := m.Clone()
	removeNodes(&working, delta.RemovedNodes)
	removeRelations(&working, delta.RemovedRelations)
	working.Nodes = append(working.Nodes, delta.AddedNodes...)
	working.Relations = append(working.Relations, delta.AddedRelations...)
	working.Normalize()
	if err := working.Validate(); err != nil {
		return Model{}, err
	}
	return working, nil
}

func removeNodes(model *Model, nodes []Node) {
	removed := make(map[ID]struct{}, len(nodes))
	for _, node := range nodes {
		removed[node.ID] = struct{}{}
	}
	kept := model.Nodes[:0]
	for _, node := range model.Nodes {
		if _, exists := removed[node.ID]; !exists {
			kept = append(kept, node)
		}
	}
	model.Nodes = kept
	keptRelations := model.Relations[:0]
	for _, relation := range model.Relations {
		if _, exists := removed[relation.Source]; exists {
			continue
		}
		if _, exists := removed[relation.Target]; exists {
			continue
		}
		keptRelations = append(keptRelations, relation)
	}
	model.Relations = keptRelations
}

func removeRelations(model *Model, relations []Relation) {
	removed := make(map[string]struct{}, len(relations))
	for _, relation := range relations {
		removed[relationKey(relation.Kind, relation.Source, relation.Target)] = struct{}{}
	}
	kept := model.Relations[:0]
	for _, relation := range model.Relations {
		if _, exists := removed[relationKey(relation.Kind, relation.Source, relation.Target)]; !exists {
			kept = append(kept, relation)
		}
	}
	model.Relations = kept
}

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
