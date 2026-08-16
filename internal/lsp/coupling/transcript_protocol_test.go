package coupling

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestProtocolTranscriptUpstreamUnknownAndFailClosed(t *testing.T) {
	tests := []struct {
		name        string
		status      Outcome
		reason      Reason
		wantOutcome Outcome
		wantCode    string
	}{
		{name: "unknown", status: OutcomeUnknown, reason: ReasonUpstreamUnknown, wantOutcome: OutcomeUnknown, wantCode: DiagnosticUpstreamUnknown},
		{name: "fail-closed", status: OutcomeFailClosed, reason: ReasonUpstreamFail, wantOutcome: OutcomeFailClosed, wantCode: DiagnosticUpstreamFail},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
			envelope.Status, envelope.Reason = test.status, test.reason
			result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
			if result.Outcome != test.wantOutcome || len(result.Links) != 0 || result.Hover != nil {
				t.Fatalf("upstream result = %#v", result)
			}
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != test.wantCode {
				t.Fatalf("upstream diagnostics = %#v", result.Diagnostics)
			}
		})
	}
}

func TestProtocolTranscriptRootRelocationAndLabelChanges(t *testing.T) {
	oldEnvelope := transcriptEnvelope("file:///old-root/main.go", "Old label", ClaimNoDelta)
	oldEnvelope.Explanations[0].Target.URI = "file:///old-root/model.gooo"
	oldEvidence, err := ComputeEvidenceDigest(oldEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	old := transcriptAdapter(t, oldEnvelope).Resolve(transcriptRequest(oldEnvelope))

	newEnvelope := transcriptEnvelope("file:///new-root/main.go", "Renamed label", ClaimNoDelta)
	newEnvelope.Explanations[0].Target.URI = "file:///new-root/model.gooo"
	newEvidence, err := ComputeEvidenceDigest(newEnvelope)
	if err != nil {
		t.Fatal(err)
	}
	if oldEvidence != newEvidence {
		t.Fatalf("presentation-only root relocation changed evidence digest: %q != %q", oldEvidence, newEvidence)
	}
	new := transcriptAdapter(t, newEnvelope).Resolve(transcriptRequest(newEnvelope))
	if old.Outcome != OutcomePass || new.Outcome != OutcomePass {
		t.Fatalf("relocation results = %#v %#v", old, new)
	}
	if old.Links[0].TargetURI == new.Links[0].TargetURI || new.Links[0].TargetURI != "file:///new-root/model.gooo" {
		t.Fatalf("root relocation target = %q from %q", new.Links[0].TargetURI, old.Links[0].TargetURI)
	}

	labelEnvelope := transcriptEnvelope("file:///old-root/main.go", "Renamed label", ClaimNoDelta)
	labelEnvelope.Explanations[0].Target.URI = oldEnvelope.Explanations[0].Target.URI
	labelResult := transcriptAdapter(t, labelEnvelope).Resolve(transcriptRequest(labelEnvelope))
	if !reflect.DeepEqual(old.Links, labelResult.Links) {
		t.Fatalf("label change altered exact navigation: old=%#v new=%#v", old.Links, labelResult.Links)
	}
	if old.Hover.Contents.Value == labelResult.Hover.Contents.Value {
		t.Fatal("label change did not alter presentation hover")
	}
}

func TestProtocolTranscriptStandardClientJSONRoundTrip(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	linksJSON, err := json.Marshal(result.Links)
	if err != nil {
		t.Fatal(err)
	}
	var links []LocationLink
	if err := json.Unmarshal(linksJSON, &links); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Links, links) {
		t.Fatalf("location links changed after JSON roundtrip: %#v %#v", result.Links, links)
	}
	allJSON, err := json.Marshal(struct {
		Links       []LocationLink `json:"links"`
		Hover       *Hover         `json:"hover"`
		Diagnostics []Diagnostic   `json:"diagnostics"`
	}{result.Links, result.Hover, result.Diagnostics})
	if err != nil {
		t.Fatal(err)
	}
	wire := string(allJSON)
	for _, forbidden := range []string{"stable_id", "code_symbol_id", "semantic_owner_id", "customWireID"} {
		if strings.Contains(wire, forbidden) {
			t.Fatalf("standard LSP wire contains custom identity field %q: %s", forbidden, wire)
		}
	}
	if !strings.Contains(wire, "relatedInformation") || !strings.Contains(wire, "targetUri") {
		t.Fatalf("standard LSP fields missing from wire: %s", wire)
	}
}

func TestProtocolTranscriptZeroWritesAndInputImmutability(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	input := transcriptBytes(t, envelope)
	original := append([]byte(nil), input...)
	adapter, err := New(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] ^= 1
	if result := adapter.Resolve(transcriptRequest(envelope)); result.Outcome != OutcomePass {
		t.Fatalf("input mutation changed result: %#v", result)
	}
	if !bytes.Equal(adapter.RawBytes(), original) {
		t.Fatal("adapter retained caller-owned mutable bytes")
	}
	if bytes.Equal(input, original) {
		t.Fatal("test did not mutate caller input")
	}
}

func TestProtocolTranscriptMissingMandatoryInputs(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimNoDelta)
	adapter := transcriptAdapter(t, envelope)
	request := transcriptRequest(envelope)
	request.Context = nil
	if result := adapter.Resolve(request); result.Diagnostics[0].Code != DiagnosticMissingCancellation || len(result.Links) != 0 {
		t.Fatalf("missing cancellation result = %#v", result)
	}
	request = transcriptRequest(envelope)
	request.SnapshotDigest = ""
	if result := adapter.Resolve(request); result.Diagnostics[0].Code != DiagnosticMissingSnapshot || len(result.Links) != 0 {
		t.Fatalf("missing snapshot result = %#v", result)
	}
	request = transcriptRequest(envelope)
	request.DocumentVersion = 0
	if result := adapter.Resolve(request); result.Diagnostics[0].Code != DiagnosticMissingVersion || len(result.Links) != 0 {
		t.Fatalf("missing version result = %#v", result)
	}
}

func TestStrictDecoderRejectsRecursiveDuplicatesAndTrailingBytes(t *testing.T) {
	fixtures := [][]byte{
		[]byte(`{"outer":{"key":1,"key":2}}`),
		[]byte(`{"schema":"gooo/lsp-coupling-explanation/v1"} {}`),
		[]byte(`{"schema":"gooo/lsp-coupling-explanation/v1"} trailing`),
	}
	for index, fixture := range fixtures {
		if _, err := New(fixture); err == nil {
			t.Fatalf("fixture %d was accepted", index)
		}
	}
}

func TestProtocolTranscriptUpstreamAndSourceMapDigestBinding(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	original := transcriptBytes(t, envelope)
	evidenceDigest, err := ComputeEvidenceDigest(envelope)
	if err != nil {
		t.Fatal(err)
	}
	envelope.EvidenceDigest = evidenceDigest

	mutated := envelope
	mutated.RegistryDigest = transcriptDigest("different-registry")
	if _, err := New(marshalWithoutRebindingEvidence(t, mutated)); err == nil {
		t.Fatal("registry digest mismatch was accepted")
	}
	mutated = envelope
	mutated.DetectorResultDigest = transcriptDigest("different-detector")
	if _, err := New(marshalWithoutRebindingEvidence(t, mutated)); err == nil {
		t.Fatal("detector result digest mismatch was accepted")
	}
	mutated = envelope
	mutated.Explanations[0].Origin.SourceMapDigest = transcriptDigest("different-source-map")
	if _, err := New(marshalWithoutRebindingEvidence(t, mutated)); err == nil {
		t.Fatal("source-map digest mismatch was accepted")
	}
	mutated = envelope
	mutated.Explanations[0].CodeSymbolID = "display-name-is-not-an-identity"
	if _, err := New(marshalWithoutRebindingEvidence(t, mutated)); err == nil {
		t.Fatal("non-URI stable ID was accepted")
	}
	if len(original) == 0 {
		t.Fatal("fixture unexpectedly encoded to zero bytes")
	}
}

func marshalWithoutRebindingEvidence(t *testing.T, envelope Envelope) []byte {
	t.Helper()
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestProtocolTranscriptReadOnlyStandardSurface(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	wire, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(wire, &fields); err != nil {
		t.Fatal(err)
	}
	for field := range fields {
		switch field {
		case "Outcome", "Links", "Hover", "Diagnostics":
		default:
			t.Fatalf("read-only adapter exposed an unexpected wire field %q: %s", field, wire)
		}
	}
	if len(result.Links) != 1 || len(result.Diagnostics) != 1 {
		t.Fatalf("read-only result = %#v", result)
	}
}
