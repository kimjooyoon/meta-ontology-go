package claimledger

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func evidenceMatches(spec EvidenceSpec, observed any, subject string) (bool, string) {
	if spec.Operator == "NON_NULL" {
		return observed != nil, ""
	}
	if spec.Operator == "POSITIVE_INTEGER" {
		number, ok := observed.(json.Number)
		if !ok {
			return false, ""
		}
		value, err := number.Int64()
		return err == nil && value > 0, ""
	}
	decoder := json.NewDecoder(bytes.NewReader(spec.Expected))
	decoder.UseNumber()
	var expected any
	if err := decoder.Decode(&expected); err != nil {
		return false, "invalid-expected-value"
	}
	if text, ok := expected.(string); ok && text == "$SUBJECT_SHA" {
		expected = subject
	}
	observedJSON, observedErr := json.Marshal(observed)
	expectedJSON, expectedErr := json.Marshal(expected)
	matched := observedErr == nil && expectedErr == nil && bytes.Equal(observedJSON, expectedJSON)
	return matched, digestValue(expected)
}

func digestBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "unencodable"
	}
	return digestBytes(encoded)
}
