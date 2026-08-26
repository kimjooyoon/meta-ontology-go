package workfrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

func canonicalR4PayloadDigest(payload string) (string, error) {
	raw := []byte(payload)
	if len(raw) == 0 || !utf8.Valid(raw) {
		return "", fmt.Errorf("payload is not valid UTF-8")
	}
	if err := rejectR4DuplicateKeys(raw); err != nil {
		return "", fmt.Errorf("payload object: %w", err)
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("payload JSON: %w", err)
	}
	if value == nil {
		return "", fmt.Errorf("payload must be a JSON object")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return "", fmt.Errorf("canonical payload: %w", err)
	}
	if !bytes.Equal(compact.Bytes(), raw) {
		return "", fmt.Errorf("payload is not canonical")
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}
