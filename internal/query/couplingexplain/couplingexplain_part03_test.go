package couplingexplain

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestStrictEnvelopeJSONAndAdapterBoundary(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeVerifiedEnvelope(raw)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Digest() != envelope.Digest() {
		t.Fatalf("decoded digest = %q, original = %q", decoded.Digest(), envelope.Digest())
	}
	if _, err := DecodeVerifiedEnvelope(append(raw, []byte(` {}`)...)); err == nil {
		t.Fatal("trailing JSON accepted")
	}
	unknown := bytes.Replace(raw, []byte(`"schema"`), []byte(`"unknown":true,"schema"`), 1)
	if _, err := DecodeVerifiedEnvelope(unknown); err == nil {
		t.Fatal("unknown JSON field accepted")
	}
	topLevelDuplicate := bytes.Replace(raw, []byte(`"schema":"gooo-coupling-explanation/v1","binding"`), []byte(`"schema":"gooo-coupling-explanation/v1","schema":"gooo-coupling-explanation/v1","binding"`), 1)
	if _, err := DecodeVerifiedEnvelope(topLevelDuplicate); err == nil {
		t.Fatal("top-level duplicate JSON key accepted")
	}
	nestedDuplicate := bytes.Replace(raw, []byte(`"snapshot_digest":"`+request.SnapshotDigest+`","registry_digest"`), []byte(`"snapshot_digest":"`+request.SnapshotDigest+`","snapshot_digest":"`+request.SnapshotDigest+`","registry_digest"`), 1)
	if _, err := DecodeVerifiedEnvelope(nestedDuplicate); err == nil {
		t.Fatal("nested duplicate JSON key accepted")
	}
	adapter := fixtureAdapter{}
	got, err := ExplainWithAdapter(context.Background(), request, raw, adapter)
	if err != nil || got.Status != StatusPass || got.Link == nil {
		t.Fatalf("adapter result = %#v, err=%v", got, err)
	}
}
func TestMissingAdaptersAreUnknownWithoutLink(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
	for _, name := range []string{"detector", "canonical"} {
		t.Run(name, func(t *testing.T) {
			var got Explanation
			var err error
			if name == "detector" {
				got, err = ExplainDetectorSnapshot(context.Background(), request, DetectorSnapshot{}, nil)
			} else {
				got, err = ExplainCanonicalSnapshot(context.Background(), request, CanonicalInputs{}, nil)
			}
			if err != nil || got.Status != StatusUnknown || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != ReasonMissing {
				t.Fatalf("missing adapter result = %#v, err=%v", got, err)
			}
		})
	}
	if envelope.EnvelopeDigest == "" {
		t.Fatal("fixture envelope digest missing")
	}
}
