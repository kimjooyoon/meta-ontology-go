package semantic

// ApplyGraphPatch validates and applies a typed mutation to a deep graph copy.
func (g Graph) ApplyGraphPatch(base GraphPatchBase, request GraphPatchRequest, mutation GraphPatchMutation) (Graph, error) {
	if err := g.Validate(); err != nil {
		return Graph{}, patchConflict(PatchInvalidRequest, "graph: "+err.Error())
	}
	if err := g.ValidatePatchPreconditions(base, request); err != nil {
		return Graph{}, err
	}
	clone := g.Clone()
	switch request.Operation {
	case GraphPatchSetNodeField:
		if err := clone.applyNodeMutation(request, mutation); err != nil {
			return Graph{}, err
		}
	case GraphPatchAddFact:
		if err := clone.applyFactMutation(request, mutation); err != nil {
			return Graph{}, err
		}
	default:
		return Graph{}, patchConflict(PatchInvalidRequest, "unsupported operation")
	}
	if err := clone.Validate(); err != nil {
		return Graph{}, patchConflict(PatchInvalidRequest, "resulting graph: "+err.Error())
	}
	return clone, nil
}
func (g *Graph) applyNodeMutation(request GraphPatchRequest, mutation GraphPatchMutation) error {
	id, err := ParseIdentity(request.NodeID.String())
	if err != nil {
		return patchConflict(PatchInvalidRequest, "node ID: "+err.Error())
	}
	node, ok := g.Node(id)
	if !ok {
		return patchConflict(PatchUnknownNode, id.String())
	}
	switch request.Field {
	case "name":
		node.Name = mutation.Name
	case "aliases":
		node.Aliases = append([]string(nil), mutation.Aliases...)
	default:
		return patchConflict(PatchInvalidRequest, "unsupported mutation field")
	}
	if err := g.AddNode(node); err != nil {
		return patchConflict(PatchInvalidRequest, "node mutation: "+err.Error())
	}
	return nil
}
func (g *Graph) applyFactMutation(request GraphPatchRequest, mutation GraphPatchMutation) error {
	if mutation.Fact == nil {
		return patchConflict(PatchInvalidRequest, "fact mutation is required")
	}
	fact, err := mutation.Fact.Normalized()
	if err != nil {
		return patchConflict(PatchInvalidRequest, "fact mutation: "+err.Error())
	}
	if fact.Status != FactDeterministic || fact.Subject != request.Subject ||
		fact.Predicate != request.Predicate || fact.Object != request.Object {
		return patchConflict(PatchInvalidRequest, "fact mutation does not match request")
	}
	if err := g.AddFact(fact); err != nil {
		return patchConflict(PatchInvalidRequest, "fact mutation: "+err.Error())
	}
	return nil
}
