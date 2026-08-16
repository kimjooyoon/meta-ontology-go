package workfrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

type r4Binding struct {
	payload string
	digest  string
	reason  string
}

func validateR4Bindings(input R4Input) string {
	bindings := []r4Binding{
		{payload: input.SnapshotPayload, digest: input.SnapshotDigest, reason: R4ReasonSnapshotBindingMismatch},
		{payload: input.PolicyPayload, digest: input.PolicyDigest, reason: R4ReasonPolicyBindingMismatch},
		{payload: input.RegistryPayload, digest: input.RegistryDigest, reason: R4ReasonRegistryBindingMismatch},
	}
	for _, binding := range bindings {
		computed, err := canonicalR4PayloadDigest(binding.payload)
		if err != nil {
			return R4ReasonMalformedBinding
		}
		if computed != binding.digest {
			return binding.reason
		}
	}
	return ""
}

func canonicalR4PayloadDigest(payload string) (string, error) {
	raw := []byte(payload)
	if len(raw) == 0 || !utf8.Valid(raw) {
		return "", fmt.Errorf("payload is not valid UTF-8")
	}
	if err := rejectR4DuplicateKeys(raw); err != nil {
		return "", fmt.Errorf("payload object: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("payload JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return "", fmt.Errorf("payload has multiple values")
		}
		return "", fmt.Errorf("payload JSON: %w", err)
	}
	if _, ok := value.(map[string]any); !ok {
		return "", fmt.Errorf("payload must be a JSON object")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("canonical payload: %w", err)
	}
	if !bytes.Equal(canonical, raw) {
		return "", fmt.Errorf("payload is not canonical")
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}
