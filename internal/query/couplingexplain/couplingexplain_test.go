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
	envelope.Receipt.EvidenceRefs = []string{"evidence://z", "evidence://a"}
	envelope.Verifier.EvidenceRefs = []string{"path://z", "path://a"}
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
	adapter := fixtureAdapter{}
	got, err := ExplainWithAdapter(context.Background(), request, raw, adapter)
	if err != nil || got.Status != StatusPass || got.Link == nil {
		t.Fatalf("adapter result = %#v, err=%v", got, err)
	}
}

func TestMergedDetectorResultIsStrictAndAdapterOnly(t *testing.T) {
	result := coupling.Evaluate(coupling.Input{}, coupling.AuthorityContext{})
	if result.Status != coupling.StatusUnknown || len(result.Reasons) != 1 || result.Reasons[0].Code != coupling.ReasonAuthorityInputSelfBound {
		t.Fatalf("missing evaluator authority = %#v", result)
	}
	raw, err := coupling.EncodeResult(result)
	if err != nil {
		t.Fatal(err)
	}
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

type fixtureAdapter struct{}

func (fixtureAdapter) DecodeVerifiedEnvelope(data []byte) (VerifiedEnvelope, error) {
	return DecodeVerifiedEnvelope(data)
}

type detectorFixtureAdapter struct {
	envelope VerifiedEnvelope
}

func (adapter detectorFixtureAdapter) AdaptDetectorSnapshot(DetectorSnapshot) (VerifiedEnvelope, error) {
	return adapter.envelope, nil
}

func fixtureEnvelope(t *testing.T, claim ChangeClaim, verdict EnvelopeVerdict) (Request, VerifiedEnvelope) {
	t.Helper()
	bindingDigest := digest("binding")
	termDigest := digest("term-definition")
	evidenceDigest := digest("evidence")
	control := Control{RequestVersion: 7, ObservedVersion: 7, RequestCancellationVersion: 11, ObservedCancellationVersion: 11}
	binding := SnapshotBinding{SnapshotDigest: digest("snapshot"), RegistryDigest: digest("registry"), SourceMapDigest: digest("source-map"), ManifestDigest: digest("manifest"), ToolchainDigest: digest("toolchain"), ProfileDigest: digest("profile"), DetectorInputDigest: digest("detector-input"), DetectorResultDigest: digest("detector-result"), VerifierResultDigest: digest("verifier-result"), Control: control}
	path := PathSummary{PathID: "path://pay-order", StartID: "code://billing/pay-order", EndID: "evidence://coupling", StepCount: 3, PathDigest: digest("path"), Steps: []PathStep{
		{FromID: "code://billing/pay-order", ToID: "owner://billing/pay-order", Kind: semantic.InferenceDerivedProjection, Phase: semantic.PhasePlacement{Phase: semantic.PhaseProjection, Ordinal: 1}, InputDigest: bindingDigest, OutputDigest: digest("owner")},
		{FromID: "owner://billing/pay-order", ToID: "term://pay-order", Kind: semantic.InferenceAuthoritativeDeclaration, Phase: semantic.PhasePlacement{Phase: semantic.PhaseDeclaration, Ordinal: 2}, InputDigest: digest("owner"), OutputDigest: termDigest},
		{FromID: "term://pay-order", ToID: "evidence://coupling", Kind: semantic.InferenceIndependentVerification, Phase: semantic.PhasePlacement{Phase: semantic.PhaseVerification, Ordinal: 3}, InputDigest: termDigest, OutputDigest: evidenceDigest, EvidenceRef: "evidence://coupling"},
	}}
	receipt := ReceiptSummary{ReceiptID: "receipt://pay-order", SurfaceID: "surface://pay-order", ChangeClaim: claim, ReceiptKind: semantic.SemanticDelta, BeforeIRDigest: digest("before"), AfterIRDigest: digest("after"), DeltaDigest: digest("delta"), OriginPathID: path.PathID, EvidenceRefs: []string{"evidence://coupling"}}
	if claim == ClaimNoDelta {
		receipt.ReceiptKind = semantic.NoSemanticDelta
		receipt.AfterIRDigest = receipt.BeforeIRDigest
		receipt.DeltaDigest = ""
	}
	verifier := VerifierSummary{EvidenceID: "evidence://coupling", ReceiptID: receipt.ReceiptID, State: VerifierPass, Independent: true, EvidenceDigest: evidenceDigest, EvidenceRefs: []string{path.PathID}}
	envelope := VerifiedEnvelope{Schema: "gooo-coupling-explanation/v1", Binding: binding, CodeBinding: CodeBindingSummary{CodeSymbolID: "code://billing/pay-order", SemanticOwnerID: "owner://billing/pay-order", RegisteredSurfaceID: receipt.SurfaceID, SourceMapID: "sourcemap://pay-order", BindingDigest: bindingDigest}, SemanticOwner: "owner://billing/pay-order", Term: TermSummary{TermID: "term://pay-order", SemanticOwnerID: "owner://billing/pay-order", Version: "v1", DefinitionDigest: termDigest}, OriginPath: path, Receipt: receipt, Verifier: verifier, Verdict: verdict, EvidenceDigest: evidenceDigest}
	if verdict != VerdictVerified {
		envelope.NoLinkReason = ReasonMissing
		envelope.Diagnostics = []Diagnostic{{Code: "fixture-no-link"}}
	}
	envelope.EnvelopeDigest = envelope.Digest()
	request := Request{CodeSymbolID: envelope.CodeBinding.CodeSymbolID, SnapshotDigest: binding.SnapshotDigest, RegistryDigest: binding.RegistryDigest, SourceMapDigest: binding.SourceMapDigest, ManifestDigest: binding.ManifestDigest, ToolchainDigest: binding.ToolchainDigest, ProfileDigest: binding.ProfileDigest, DetectorInputDigest: binding.DetectorInputDigest, DetectorResultDigest: binding.DetectorResultDigest, VerifierResultDigest: binding.VerifierResultDigest, EnvelopeDigest: envelope.EnvelopeDigest, Control: control}
	return request, envelope
}

func refreshEnvelopeDigest(request *Request, envelope *VerifiedEnvelope) {
	envelope.EnvelopeDigest = envelope.Digest()
	request.EnvelopeDigest = envelope.EnvelopeDigest
}

func digest(value string) string { return DigestBytes([]byte(value)) }

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestNoLinkReasonsAreClosed(t *testing.T) {
	for _, value := range []LinkReason{ReasonAmbiguous, ReasonStale, ReasonUnregistered, ReasonMissing, ReasonNotVerified} {
		if !validReason(value) {
			t.Fatal(value)
		}
	}
}
