package couplingexplain

import (
	"bytes"
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"testing"
)

func TestMergedDetectorResultIsStrictAndAdapterOnly(t *testing.T) {
	result := literalDetectorResult()
	if result.Status != coupling.StatusUnknown || len(result.Reasons) != 1 || result.Reasons[0].Code != coupling.ReasonAuthorityInputSelfBound {
		t.Fatalf("missing evaluator authority = %#v", result)
	}
	raw := literalDetectorResultBytes()
	decoded, err := DecodeDetectorResult(raw)
	if err != nil || decoded.Digest != result.Digest || decoded.Status != coupling.StatusUnknown {
		t.Fatalf("detector result = %#v, err=%v", decoded, err)
	}
	if _, err := DecodeDetectorResult(append(raw, []byte(` {}`)...)); err == nil {
		t.Fatal("detector result trailing JSON accepted")
	}
	tampered := bytes.Replace(raw, []byte(result.Digest), []byte(digest("tampered-detector-result")), 1)
	if _, err := DecodeDetectorResult(tampered); err == nil {
		t.Fatal("detector result digest tampering accepted")
	}
	unknown := bytes.Replace(raw, []byte(`"schema"`), []byte(`"unknown":true,"schema"`), 1)
	if _, err := DecodeDetectorResult(unknown); err == nil {
		t.Fatal("detector result unknown field accepted")
	}
	request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
	got, err := ExplainDetectorSnapshot(context.Background(), request, DetectorSnapshot{InputBytes: []byte("input"), ResultBytes: raw}, detectorFixtureAdapter{envelope: envelope})
	if err != nil || got.Status != StatusPass || got.Link == nil {
		t.Fatalf("detector adapter result = %#v, err=%v", got, err)
	}
}
func TestProducerResultMutationsRejectExactDigest(t *testing.T) {
	cases := []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"decision", func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"status":"UNKNOWN"`), []byte(`"status":"PASS"`), 1)
		}},
		{"reason", func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"detail":"evaluator authority context is missing"`), []byte(`"detail":"changed producer reason"`), 1)
		}},
		{"input-digest", func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`265a7627c123865b1cb0a3cadfc74b0d9e079cfa85a78dfbc1534368d73c2beb`), []byte(digest("other-input")), 1)
		}},
		{"binding", func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"accepted_surface_ids":null`), []byte(`"accepted_surface_ids":["surface://mutated"]`), 1)
		}},
		{"count", func(raw []byte) []byte {
			return bytes.Replace(raw, []byte(`"receipts":{"known":false,"value":0}`), []byte(`"receipts":{"known":true,"value":1}`), 1)
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			raw := test.mutate(literalDetectorResultBytes())
			_, err := DecodeDetectorResult(raw)
			jsonErr, ok := err.(*coupling.JSONError)
			if !ok || jsonErr.Code != coupling.ReasonMalformedBinding {
				t.Fatalf("producer mutation %q error = %v, want MALFORMED_BINDING", test.name, err)
			}
			if err == nil {
				t.Fatalf("producer mutation %q accepted", test.name)
			}
		})
	}
}
