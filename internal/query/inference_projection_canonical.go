package query

import (
	"encoding/json"
)

// CanonicalJSON returns the normalized request representation. It contains
// only typed selectors and explicit bounds, so replay does not depend on map
// iteration or process state.
func (query InferenceQuery) CanonicalJSON() ([]byte, error) {
	normalized, err := query.normalized()
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

func (query InferenceQuery) Normalize() (InferenceQuery, error) { return query.normalized() }

func (query InferenceQuery) CanonicalDigest() (string, error) {
	canonical, err := query.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

// CanonicalJSON omits the self-referential hash and is stable across fresh
// processes. Rejected results are canonical receipts too.
func (result InferenceQueryResult) CanonicalJSON() ([]byte, error) {
	canonical := result
	canonical.Hash = ""
	return json.Marshal(canonical)
}

func (result InferenceQueryResult) CanonicalDigest() (string, error) {
	canonical, err := result.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

func (result InferenceQueryResult) CanonicalDigestValue() string {
	digest, _ := result.CanonicalDigest()
	return digest
}

func (result *InferenceQueryResult) seal() error {
	digest, err := result.CanonicalDigest()
	if err != nil {
		return err
	}
	result.Hash = digest
	return nil
}

func inferenceRowCanonical(row InferenceRow) string {
	canonical, _ := json.Marshal(row)
	return string(canonical)
}

func semanticChangeRowCanonical(row SemanticChangeRow) string {
	canonical, _ := json.Marshal(row)
	return string(canonical)
}
