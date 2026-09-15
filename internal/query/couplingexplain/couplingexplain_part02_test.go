package couplingexplain

import (
	"bytes"
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestExplainRejectsMalformedOrContradictoryVerifiedEnvelope(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(*VerifiedEnvelope)
		wantReason LinkReason
	}{
		{"wrong-owner", func(envelope *VerifiedEnvelope) { envelope.SemanticOwner = "owner://wrong" }, ReasonUnregistered},
		{"disconnected", func(envelope *VerifiedEnvelope) { envelope.OriginPath.Steps[1].FromID = "disconnected://node" }, ReasonAmbiguous},
		{"fork", func(envelope *VerifiedEnvelope) {
			envelope.OriginPath.Steps[1].FromID = envelope.OriginPath.Steps[0].FromID
		}, ReasonAmbiguous},
		{"cycle", func(envelope *VerifiedEnvelope) {
			envelope.OriginPath.Steps[2].ToID = envelope.OriginPath.StartID
			envelope.OriginPath.EndID = envelope.OriginPath.StartID
		}, ReasonAmbiguous},
		{"candidate", func(envelope *VerifiedEnvelope) {
			envelope.OriginPath.Steps[1].Kind = semantic.InferenceObservationCandidate
		}, ReasonAmbiguous},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
			test.mutate(&envelope)
			refreshEnvelopeDigest(&request, &envelope)
			got := Explain(context.Background(), request, envelope)
			if got.Status != StatusFailClosed || got.Link != nil || got.NoLink == nil || got.NoLink.Reason != test.wantReason {
				t.Fatalf("result = %#v, want FAIL_CLOSED/%s without link", got, test.wantReason)
			}
		})
	}
}
func TestEnvelopeLabelsRootsActorsAndPermutationsDoNotChangeEvidence(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
	envelope.CodeBinding.Presentation = Presentation{Label: "renamed", Root: "new-root", Path: "/different.go", Timestamp: "later", Actor: "agent-b"}
	envelope.Term.Presentation = Presentation{Label: "new term", Root: "term-root", Path: "term.go", Timestamp: "later", Actor: "agent-c"}
	envelope.OriginPath.Presentation = Presentation{Label: "path label", Root: "path-root", Path: "path", Timestamp: "later", Actor: "agent-d"}
	envelope.Receipt.Presentation = Presentation{Label: "receipt label", Root: "receipt-root", Path: "receipt", Timestamp: "later", Actor: "agent-e"}
	envelope.Verifier.Presentation = Presentation{Label: "verifier label", Root: "verifier-root", Path: "verifier", Timestamp: "later", Actor: "agent-f"}
	refreshEnvelopeDigest(&request, &envelope)
	got := Explain(context.Background(), request, envelope)
	if got.Status != StatusPass || got.EvidenceDigest != envelope.EvidenceDigest {
		t.Fatalf("result = %#v", got)
	}
	compact, err := got.CanonicalJSON(ViewCompact)
	if err != nil {
		t.Fatal(err)
	}
	expanded, err := got.CanonicalJSON(ViewExpanded)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(compact, expanded) || !bytes.Contains(expanded, []byte(`"steps"`)) || bytes.Contains(compact, []byte(`"steps"`)) {
		t.Fatalf("compact/expanded presentation mismatch: compact=%s expanded=%s", compact, expanded)
	}
	if bytes.Contains(compact, []byte("renamed")) || bytes.Contains(expanded, []byte("agent-f")) ||
		bytes.Contains(compact, []byte("new-root")) {
		t.Fatalf("presentation leaked into canonical output")
	}
	decoded := got
	decoded.Link.OriginPath.Steps = nil
	if decoded.EvidenceDigest != got.EvidenceDigest || got.Link.CodeBinding.SemanticOwnerID != "owner://billing/pay-order" {
		t.Fatal("view projection changed the evidence or authority tuple")
	}
}
