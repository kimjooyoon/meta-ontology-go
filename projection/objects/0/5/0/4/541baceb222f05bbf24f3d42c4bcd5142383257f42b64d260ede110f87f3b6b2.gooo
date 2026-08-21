package query

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

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
	if envelope, ok := errors.AsType[*EnvelopeError](err); ok {
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
