package bidir

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
	if err := validateFieldParentStability(m, working); err != nil {
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
