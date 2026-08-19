package query

func normalizeTraversal(request Request) (Request, error) {
	if request.Rule != "" {
		return Request{}, envelopeError(ErrInvalidEnvelope, "unexpected_rule", "traversal does not accept derived rules")
	}
	if request.Target != "" {
		return Request{}, envelopeError(ErrInvalidEnvelope, "unexpected_target", "traversal has no target")
	}
	direction, err := normalizeDirection(request.Direction)
	if err != nil {
		return Request{}, err
	}
	if direction == "" {
		return Request{}, envelopeError(
			ErrAmbiguousTraversal, "ambiguous_direction", "traversal requires outgoing, incoming, or both",
		)
	}
	if request.Relation != "" {
		relation, err := ParseRelation(request.Relation)
		if err != nil {
			return Request{}, envelopeError(ErrInvalidRelation, "unsupported_relation", err.Error())
		}
		request.Relation = relation
	}
	request.Direction = direction
	return request, nil
}
func normalizeDerived(request Request) (Request, error) {
	if request.Target != "" {
		return Request{}, envelopeError(ErrInvalidDerivedQuery, "unexpected_target", "derived queries have no target")
	}
	if request.Relation != "" {
		return Request{}, envelopeError(
			ErrInvalidDerivedQuery, "derived_relation_rejected", "derived rules use registered rule IDs",
		)
	}
	if request.Direction != "" && request.Direction != "outgoing" {
		return Request{}, envelopeError(
			ErrInvalidDerivedQuery, "reversed_direction", "derived rules have a fixed outgoing result direction",
		)
	}
	rule, err := ParseDerivedRule(request.Rule)
	if err != nil {
		return Request{}, envelopeError(ErrUnsupportedDerivedRule, "unsupported_rule", err.Error())
	}
	_, selection, err := normalizeLayer(request.Layer)
	if err != nil {
		return Request{}, err
	}
	if _, _, err := normalizeDerivedOptions(DerivedOptions{
		Rule: rule, MaxDepth: request.MaxDepth, Limit: request.Limit, Selection: selection,
	}); err != nil {
		return Request{}, envelopeError(ErrInvalidDerivedQuery, "invalid_rule_options", err.Error())
	}
	request.Rule = rule
	request.Direction = "outgoing"
	return request, nil
}
func normalizeLayer(layer Layer) (Layer, FactSelection, error) {
	switch layer {
	case LayerAll:
		return layer, SelectAll, nil
	case LayerDeterministic:
		return layer, SelectDeterministic, nil
	case LayerCandidate:
		return layer, SelectCandidate, nil
	default:
		return "", 0, envelopeError(ErrUnsupportedLayer, "unsupported_layer", string(layer))
	}
}
