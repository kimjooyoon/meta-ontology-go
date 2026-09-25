package proposalpredecessor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func validObservationCacheRaw(t *testing.T) []byte {
	t.Helper()
	raw, err := json.Marshal(ObservationCache{
		Schema: ObservationSchema,
		Responses: []ObservationResponse{{
			Kind: "GET", URL: "https://api.example.test/observed", StatusCode: 200,
			Body: []byte(`{"ok":true}`),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func validObservationEvidence(raw []byte) ObservationEvidence {
	sum := sha256.Sum256(raw)
	return ObservationEvidence{
		Schema:           ObservationSchema,
		CachePath:        ObservationMemberPath,
		CacheRole:        ObservationRole,
		CacheBytes:       len(raw),
		CacheDigest:      "sha256:" + hex.EncodeToString(sum[:]),
		ResponseTotal:    1,
		ResponseConsumed: 1,
	}
}

func TestValidateRawObservationEvidenceRejectsTampering(t *testing.T) {
	raw := validObservationCacheRaw(t)
	evidence := validObservationEvidence(raw)
	tests := []struct {
		name   string
		input  func() (ObservationEvidence, []byte, int)
		reason string
	}{
		{
			name: "cache byte tamper",
			input: func() (ObservationEvidence, []byte, int) {
				altered := bytes.Replace(raw, []byte("observed"), []byte("observex"), 1)
				return evidence, altered, 1
			},
			reason: "FAIL_CLOSED: proposal observation cache digest mismatch",
		},
		{
			name: "cache digest reseal",
			input: func() (ObservationEvidence, []byte, int) {
				altered := evidence
				altered.CacheDigest = "sha256:" + strings.Repeat("0", 64)
				return altered, raw, 1
			},
			reason: "FAIL_CLOSED: proposal observation cache digest mismatch",
		},
		{
			name: "response count tamper",
			input: func() (ObservationEvidence, []byte, int) {
				alteredRaw, err := json.Marshal(ObservationCache{Schema: ObservationSchema})
				if err != nil {
					t.Fatal(err)
				}
				altered := validObservationEvidence(alteredRaw)
				altered.ResponseTotal = 1
				return altered, alteredRaw, 1
			},
			reason: "FAIL_CLOSED: proposal observation cache response count mismatch",
		},
		{
			name: "response consumed tamper",
			input: func() (ObservationEvidence, []byte, int) {
				return evidence, raw, 0
			},
			reason: "FAIL_CLOSED: proposal observation consumed count mismatch",
		},
		{
			name: "receipt response consumed tamper",
			input: func() (ObservationEvidence, []byte, int) {
				altered := evidence
				altered.ResponseConsumed = 0
				return altered, raw, 1
			},
			reason: "FAIL_CLOSED: proposal observation evidence coverage is incomplete",
		},
		{
			name: "logical path tamper",
			input: func() (ObservationEvidence, []byte, int) {
				altered := evidence
				altered.CachePath = "other-cache.json"
				return altered, raw, 1
			},
			reason: "FAIL_CLOSED: proposal observation evidence identity is invalid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alteredEvidence, alteredRaw, consumed := test.input()
			if err := ValidateRawObservationEvidence(alteredEvidence, alteredRaw, consumed); err == nil || err.Error() != test.reason {
				t.Fatalf("got %v, want %s", err, test.reason)
			}
		})
	}
}
