package coupling

import (
	"testing"
)

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
