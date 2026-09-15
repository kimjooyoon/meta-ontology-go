package couplingexplain

import (
	"context"
	"testing"
)

func TestMissingVerifiedMaterialIsUnknown(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*VerifiedEnvelope)
	}{
		{"term-version", func(envelope *VerifiedEnvelope) { envelope.Term.Version = "" }},
		{"path", func(envelope *VerifiedEnvelope) { envelope.OriginPath.Steps = nil }},
		{"verifier", func(envelope *VerifiedEnvelope) { envelope.Verifier.EvidenceID = "" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request, envelope := fixtureEnvelope(t, ClaimNoDelta, VerdictVerified)
			test.mutate(&envelope)
			refreshEnvelopeDigest(&request, &envelope)
			got := Explain(context.Background(), request, envelope)
			if got.Status != StatusUnknown || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != ReasonMissing {
				t.Fatalf("result = %#v, want UNKNOWN/MISSING", got)
			}
			if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "missing-verified-material" {
				t.Fatalf("diagnostics = %#v", got.Diagnostics)
			}
		})
	}
}
func TestMismatchedAuthoritativeBindingIsUnknown(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
	request.SnapshotDigest = digest("other-snapshot")
	got := Explain(context.Background(), request, envelope)
	if got.Status != StatusUnknown || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != ReasonStale {
		t.Fatalf("result = %#v, want UNKNOWN/STALE", got)
	}
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "stale-snapshot-binding" {
		t.Fatalf("diagnostics = %#v", got.Diagnostics)
	}
}
