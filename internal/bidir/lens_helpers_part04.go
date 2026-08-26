package bidir

func (m Model) node(id ID) (Node, bool) {
	for _, node := range m.Nodes {
		if node.ID == id {
			return node, true
		}
	}
	return Node{}, false
}
func implicitRelationKeys(declarations []Declaration, namespace string) map[string]struct{} {
	ids := make(map[string]ID, len(declarations))
	for _, declaration := range declarations {
		if id, err := declarationIdentity(namespace, declaration); err == nil {
			ids[referenceKey(namespace, declaration.Name)] = id
		}
	}
	result := make(map[string]struct{})
	for _, declaration := range declarations {
		if declaration.Kind != ActivityKind {
			continue
		}
		activityID, err := declarationIdentity(namespace, declaration)
		if err != nil {
			continue
		}
		for _, reference := range declaration.Inputs {
			id := reference.ID
			if id == "" {
				id = ids[referenceKey(namespace, reference.Name)]
			}
			if id != "" {
				result[relationKey(PredicateUsed, activityID, id)] = struct{}{}
			}
		}
		for _, reference := range declaration.Outputs {
			id := reference.ID
			if id == "" {
				id = ids[referenceKey(namespace, reference.Name)]
			}
			if id != "" {
				result[relationKey(PredicateWasGeneratedBy, id, activityID)] = struct{}{}
			}
		}
	}
	return result
}
func relationLess(left, right Relation) bool {
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	if left.Source != right.Source {
		return left.Source < right.Source
	}
	return left.Target < right.Target
}
