package query

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

func normalizeDirection(direction string) (string, error) {
	switch direction {
	case "", "outgoing", "incoming", "both":
		return direction, nil
	default:
		return "", envelopeError(ErrUnsupportedDirection, "unsupported_direction", direction)
	}
}

func envelopeError(kind error, code, detail string) *EnvelopeError {
	return &EnvelopeError{
		Code: code, Message: fmt.Sprintf("%s: %s", kind.Error(), detail), cause: kind,
	}
}

func (request Request) CanonicalJSON() ([]byte, error) {
	normalized, err := request.Normalize()
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

// CanonicalDigest hashes the normalized request bytes.
func (request Request) CanonicalDigest() (string, error) {
	canonical, err := request.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func errorCode(err error) string {
	var envelope *EnvelopeError
	if errors.As(err, &envelope) {
		return envelope.Code
	}
	if errors.Is(err, ErrUnknownEndpoint) {
		return "unknown_endpoint"
	}
	if errors.Is(err, ErrInvalidTraversal) {
		return "invalid_traversal"
	}
	if errors.Is(err, ErrInvalidQuery) {
		return "invalid_query"
	}
	if errors.Is(err, ErrUnsupportedDerivedRule) {
		return "unsupported_rule"
	}
	if errors.Is(err, ErrInvalidDerivedQuery) {
		return "invalid_derived_query"
	}
	return "query_rejected"
}
