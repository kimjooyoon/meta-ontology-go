package semantic

import "fmt"

// ValidateWithTypes validates the graph and resolves every field TypeRef
// against the supplied nominal registry without changing the graph.
func (g Graph) ValidateWithTypes(registry TypeRegistry) error {
	if err := g.Validate(); err != nil {
		return err
	}
	for _, node := range g.Nodes() {
		for _, field := range node.Fields {
			if _, err := registry.Resolve(field.TypeRef); err != nil {
				return fmt.Errorf("field %s on %s: %w", field.ID, node.ID, err)
			}
		}
	}
	return nil
}

func (g Graph) ValidateWithTypeRegistry(registry TypeRegistry) error {
	return g.ValidateWithTypes(registry)
}

// NormalizedWithTypes returns a copy whose lookup-only TypeRef metadata has
// been resolved to stable IDs. The receiver and its original nodes are never
// mutated, including when resolution fails.
func (g Graph) NormalizedWithTypes(registry TypeRegistry) (Graph, error) {
	normalized, err := g.Normalized()
	if err != nil {
		return Graph{}, err
	}
	out := NewGraph()
	for _, node := range normalized.Nodes() {
		typed, err := normalizeTypedNode(node, registry)
		if err != nil {
			return Graph{}, err
		}
		if err := out.AddNode(typed); err != nil {
			return Graph{}, err
		}
	}
	for _, fact := range normalized.AllFacts() {
		if err := out.AddFact(fact); err != nil {
			return Graph{}, err
		}
	}
	if err := out.ValidateWithTypes(registry); err != nil {
		return Graph{}, err
	}
	return out, nil
}

func (g Graph) NormalizedWithTypeRegistry(registry TypeRegistry) (Graph, error) {
	return g.NormalizedWithTypes(registry)
}

func normalizeTypedNode(node Node, registry TypeRegistry) (Node, error) {
	fields := copyFields(node.Fields)
	for i, field := range fields {
		definition, err := registry.Resolve(field.TypeRef)
		if err != nil {
			return Node{}, fmt.Errorf("field %s type ref: %w", field.ID, err)
		}
		fields[i].TypeRef = TypeRef{ID: definition.ID}
	}
	node.Fields = fields
	return node.Normalized()
}

func (ir IR) ValidateWithTypes(registry TypeRegistry) error {
	if err := ir.Validate(); err != nil {
		return err
	}
	return ir.Graph.ValidateWithTypes(registry)
}

func (ir IR) ValidateWithTypeRegistry(registry TypeRegistry) error {
	return ir.ValidateWithTypes(registry)
}

func (ir IR) NormalizedWithTypes(registry TypeRegistry) (IR, error) {
	normalized, err := ir.Normalized()
	if err != nil {
		return IR{}, err
	}
	graph, err := normalized.Graph.NormalizedWithTypes(registry)
	if err != nil {
		return IR{}, err
	}
	normalized.Graph = graph
	return normalized, nil
}

func (ir IR) NormalizedWithTypeRegistry(registry TypeRegistry) (IR, error) {
	return ir.NormalizedWithTypes(registry)
}
