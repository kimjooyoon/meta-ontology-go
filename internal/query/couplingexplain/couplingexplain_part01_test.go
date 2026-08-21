package couplingexplain

import (
	"context"
	"testing"
)

func TestExplainPositiveDeltaAndNoDelta(t *testing.T) {
	for _, claim := range []ChangeClaim{ClaimDelta, ClaimNoDelta} {
		t.Run(string(claim), func(t *testing.T) {
			request, envelope := fixtureEnvelope(t, claim, VerdictVerified)
			got := Explain(context.Background(), request, envelope)
			if got.Status != StatusPass || got.Link == nil || got.NoLink != nil {
				t.Fatalf("result = %#v, want PASS link", got)
			}
			if got.Link.CodeBinding.SemanticOwnerID != "owner://billing/pay-order" ||
				got.Link.Term.Version != "v1" || got.Link.Receipt.ChangeClaim != claim {
				t.Fatalf("authority tuple = %#v", got.Link)
			}
			if got.EvidenceDigest != envelope.EvidenceDigest {
				t.Fatalf("evidence digest = %q, want %q", got.EvidenceDigest, envelope.EvidenceDigest)
			}
		})
	}
}
func TestExplainUnknownStatesProduceNoLink(t *testing.T) {
	cases := []struct {
		name    string
		verdict EnvelopeVerdict
		reason  LinkReason
		code    string
		want    Status
	}{
		{"ambiguity", VerdictUnknown, ReasonAmbiguous, "ambiguous-binding", StatusUnknown},
		{"stale", VerdictUnknown, ReasonStale, "stale-snapshot", StatusUnknown},
		{"unregistered", VerdictUnknown, ReasonUnregistered, "unregistered-symbol", StatusUnknown},
		{"missing", VerdictUnknown, ReasonMissing, "missing-verifier", StatusUnknown},
		{"candidate-only", VerdictUnknown, ReasonAmbiguous, "candidate-only", StatusUnknown},
		{"not-verified", VerdictFailClosed, ReasonNotVerified, "verifier-failed", StatusFailClosed},
		{"verifier-fail-closed", VerdictUnknown, ReasonNotVerified, "verifier-fail-closed", StatusFailClosed},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request, envelope := fixtureEnvelope(t, ClaimNoDelta, test.verdict)
			envelope.NoLinkReason = test.reason
			envelope.Diagnostics = []Diagnostic{{Code: test.code, IDs: []string{"stable://fixture"}}}
			envelope.EvidenceDigest = digest("evidence/" + test.name)
			envelope.Verifier.EvidenceDigest = envelope.EvidenceDigest
			if test.name == "missing" {
				envelope.Verifier.EvidenceID = ""
			}
			if test.name == "verifier-fail-closed" {
				envelope.Verifier.State = VerifierFailClosed
			}
			refreshEnvelopeDigest(&request, &envelope)
			got := Explain(context.Background(), request, envelope)
			if got.Status != test.want || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != test.reason {
				t.Fatalf("result = %#v, want %s/%s without link", got, test.want, test.reason)
			}
		})
	}
}
