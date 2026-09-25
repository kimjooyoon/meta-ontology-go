package query

import (
	"fmt"
)

// Normalize validates and canonicalizes a request without touching a graph.
func (request Request) Normalize() (Request, error) {
	if request.Schema != QueryEnvelopeSchema {
		return Request{}, envelopeError(ErrInvalidEnvelope, "invalid_schema", "schema must be gooo-query/v1")
	}
	if request.Operation != OperationExact && request.Operation != OperationTraversal &&
		request.Operation != OperationDerived {
		return Request{}, envelopeError(ErrUnsupportedOperation, "unsupported_operation", string(request.Operation))
	}
	root, err := ParseID(request.Root.String())
	if err != nil {
		return Request{}, envelopeError(ErrInvalidEnvelope, "invalid_root", err.Error())
	}
	layer, _, err := normalizeLayer(request.Layer)
	if err != nil {
		return Request{}, err
	}
	if request.MaxDepth < 1 || request.MaxDepth > MaxEnvelopeDepth {
		return Request{}, envelopeError(
			ErrInvalidEnvelopeDepth, "invalid_max_depth", fmt.Sprintf("must be 1..%d", MaxEnvelopeDepth),
		)
	}
	if request.Limit < 1 || request.Limit > MaxEnvelopeLimit {
		return Request{}, envelopeError(
			ErrInvalidEnvelopeLimit, "invalid_limit", fmt.Sprintf("must be 1..%d", MaxEnvelopeLimit),
		)
	}
	request.Root = ID(root)
	request.Layer = layer
	if request.Operation == OperationExact {
		return normalizeExact(request)
	}
	if request.Operation == OperationDerived {
		return normalizeDerived(request)
	}
	return normalizeTraversal(request)
}
func normalizeExact(request Request) (Request, error) {
	if request.Rule != "" {
		return Request{}, envelopeError(ErrInvalidEnvelope, "unexpected_rule", "exact queries do not accept derived rules")
	}
	if request.Target == "" {
		return Request{}, envelopeError(ErrInvalidEnvelope, "missing_target", "exact query requires target")
	}
	target, err := ParseID(request.Target.String())
	if err != nil {
		return Request{}, envelopeError(ErrInvalidEnvelope, "invalid_target", err.Error())
	}
	relation, err := ParseRelation(request.Relation)
	if err != nil {
		return Request{}, envelopeError(ErrInvalidRelation, "unsupported_relation", err.Error())
	}
	if request.Direction != "" && request.Direction != "outgoing" {
		return Request{}, envelopeError(
			ErrAmbiguousTraversal, "ambiguous_direction", "exact queries are directed root-to-target",
		)
	}
	request.Target = ID(target)
	request.Relation = relation
	request.Direction = "outgoing"
	return request, nil
}
