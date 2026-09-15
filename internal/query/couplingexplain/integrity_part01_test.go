package couplingexplain

import (
	"context"
	"testing"
)

func TestVerifiedDigestAndEvidenceMutationsFailClosed(t *testing.T) {
	cases := []struct {
		name   string
		code   string
		mutate func(*VerifiedEnvelope)
	}{
		{"code-binding", "code-binding-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.CodeBinding.CodeBindingDigest = digest("mutated-code-binding")
		}},
		{"term", "term-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.Term.DefinitionDigest = digest("mutated-term")
		}},
		{"path", "path-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.OriginPath.PathDigest = digest("mutated-path")
		}},
		{"path-evidence-binding", "path-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.OriginPath.Steps[2].EvidenceRef = "evidence://other"
		}},
		{"receipt", "receipt-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.Receipt.ReceiptDigest = digest("mutated-receipt")
		}},
		{"receipt-evidence-binding", "receipt-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.Receipt.EvidenceRefs = []string{"evidence://other"}
		}},
		{"verifier", "verifier-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.Verifier.VerifierDigest = digest("mutated-verifier")
		}},
		{"verifier-evidence-binding", "verifier-digest-mismatch", func(envelope *VerifiedEnvelope) {
			envelope.Verifier.EvidenceID = "evidence://other"
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
			test.mutate(&envelope)
			refreshEnvelopeDigest(&request, &envelope)
			got := Explain(context.Background(), request, envelope)
			if got.Status != StatusFailClosed || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != ReasonNotVerified {
				t.Fatalf("result = %#v, want FAIL_CLOSED/NOT_VERIFIED", got)
			}
			if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != test.code {
				t.Fatalf("diagnostics = %#v, want %s", got.Diagnostics, test.code)
			}
		})
	}
}
