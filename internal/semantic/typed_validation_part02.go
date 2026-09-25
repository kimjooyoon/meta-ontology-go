package semantic

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
