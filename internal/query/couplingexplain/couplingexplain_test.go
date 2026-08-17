package couplingexplain

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestCancellationVersionRaceAndNoWrite(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
	original := append([]byte(nil), mustJSON(t, envelope)...)
	raceRequest := request
	raceRequest.Control.ObservedVersion++
	got := Explain(context.Background(), raceRequest, envelope)
	if got.Status != StatusUnknown || got.NoLink == nil || got.NoLink.Reason != ReasonStale || got.Link != nil {
		t.Fatalf("version race = %#v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got = Explain(ctx, request, envelope)
	if got.Status != StatusUnknown || got.NoLink == nil || got.NoLink.Reason != ReasonStale || got.Link != nil {
		t.Fatalf("cancellation = %#v", got)
	}
	if !bytes.Equal(original, mustJSON(t, envelope)) {
		t.Fatal("Explain mutated its envelope input")
	}
}
