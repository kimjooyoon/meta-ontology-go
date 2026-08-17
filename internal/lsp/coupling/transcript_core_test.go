package coupling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func transcriptDigest(value string) string {
	digest := sha256.Sum256([]byte("lsp-coupling/" + value))
	return hex.EncodeToString(digest[:])
}

func transcriptEnvelope(uri, label string, claim ChangeClaim) Envelope {
	return Envelope{
		Schema: SchemaVersion, SnapshotDigest: transcriptDigest("snapshot"),
		RegistryDigest: transcriptDigest("registry"), ToolchainDigest: transcriptDigest("toolchain"),
		ProfileDigest: transcriptDigest("profile"), DetectorResultDigest: transcriptDigest("detector"),
		OracleResultDigest: transcriptDigest("oracle"), Document: Document{URI: uri, Version: 7},
		Status: OutcomePass, Explanations: []Explanation{{
			CodeSymbolID: "stable://code/pay-order", SemanticOwnerID: "stable://activity/pay-order", Label: label,
			Origin: BoundLocation{
				StableID: "stable://span/code-pay-order", SourceMapID: "stable://map/code",
				SourceMapDigest: transcriptDigest("source-map-code"), URI: uri,
				Range: Range{Start: Position{Line: 4, Character: 8}, End: Position{Line: 4, Character: 17}},
			},
			Target: BoundLocation{
				StableID: "stable://span/semantic-pay-order", SourceMapID: "stable://map/semantic",
				SourceMapDigest: transcriptDigest("source-map-semantic"), URI: "file:///workspace/model.gooo",
				Range: Range{Start: Position{Line: 8, Character: 9}, End: Position{Line: 8, Character: 17}},
			},
			CausalSpans: []CausalSpan{
				{StableID: "stable://span/source", SourceMapID: "stable://map/source", SourceMapDigest: transcriptDigest("source-map-source"), URI: uri, Range: Range{Start: Position{Line: 1}, End: Position{Line: 1, Character: 10}}, Ordinal: 2, Message: "authoritative source"},
				{StableID: "stable://span/registry", SourceMapID: "stable://map/registry", SourceMapDigest: transcriptDigest("source-map-registry"), URI: "file:///workspace/registry.gooo", Range: Range{Start: Position{Line: 2}, End: Position{Line: 2, Character: 8}}, Ordinal: 1, Message: "registered binding"},
			},
			Claim: claim, Status: OutcomePass,
		}},
	}
}

func transcriptBytes(t *testing.T, envelope Envelope) []byte {
	t.Helper()
	digest, err := ComputeEvidenceDigest(envelope)
	if err != nil {
		t.Fatal(err)
	}
	envelope.EvidenceDigest = digest
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func transcriptAdapter(t *testing.T, envelope Envelope) *Adapter {
	t.Helper()
	adapter, err := New(transcriptBytes(t, envelope))
	if err != nil {
		t.Fatal(err)
	}
	return adapter
}

func transcriptRequest(envelope Envelope) Request {
	return Request{
		Context: context.Background(), DocumentURI: envelope.Document.URI, DocumentVersion: envelope.Document.Version,
		Position: Position{Line: 4, Character: 10}, SnapshotDigest: envelope.SnapshotDigest,
	}
}

func TestProtocolTranscriptDeltaAndNoDelta(t *testing.T) {
	for _, claim := range []ChangeClaim{ClaimDelta, ClaimNoDelta} {
		t.Run(string(claim), func(t *testing.T) {
			envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", claim)
			result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
			if result.Outcome != OutcomePass || len(result.Links) != 1 || result.Hover == nil {
				t.Fatalf("result = %#v, want one standard link and hover", result)
			}
			if got := result.Links[0].TargetURI; got != "file:///workspace/model.gooo" {
				t.Fatalf("target URI = %q", got)
			}
			if !strings.Contains(result.Hover.Contents.Value, string(claim)) {
				t.Fatalf("hover = %#v, missing claim %q", result.Hover, claim)
			}
			if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].RelatedInformation) != 2 {
				t.Fatalf("diagnostics = %#v, want two causal spans", result.Diagnostics)
			}
			if got := result.Diagnostics[0].RelatedInformation[0].Message; got != "registered binding" {
				t.Fatalf("first causal message = %q, ordinal order was not preserved", got)
			}
		})
	}
}

func TestProtocolTranscriptAmbiguityHasNoNavigation(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	duplicate := envelope.Explanations[0]
	duplicate.CodeSymbolID = "stable://code/other"
	duplicate.SemanticOwnerID = "stable://activity/other"
	duplicate.Origin.StableID = "stable://span/other-origin"
	duplicate.Origin.SourceMapID = "stable://map/other-code"
	duplicate.Origin.SourceMapDigest = transcriptDigest("source-map-other-code")
	duplicate.Origin.Range.End.Character = 21
	envelope.Explanations = append(envelope.Explanations, duplicate)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Hover != nil {
		t.Fatalf("ambiguous result = %#v, want no navigation", result)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != DiagnosticAmbiguous || len(result.Diagnostics[0].RelatedInformation) != 0 {
		t.Fatalf("ambiguous diagnostics = %#v", result.Diagnostics)
	}
}

func TestProtocolTranscriptStaleSnapshotAndWrongVersion(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimNoDelta)
	adapter := transcriptAdapter(t, envelope)
	stale := transcriptRequest(envelope)
	stale.SnapshotDigest = transcriptDigest("new-snapshot")
	result := adapter.Resolve(stale)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticStaleSnapshot {
		t.Fatalf("stale result = %#v", result)
	}
	wrongVersion := transcriptRequest(envelope)
	wrongVersion.DocumentVersion++
	result = adapter.Resolve(wrongVersion)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticWrongVersion {
		t.Fatalf("wrong-version result = %#v", result)
	}
}

func TestProtocolTranscriptCancellationWinsOrdering(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := transcriptRequest(envelope)
	request.Context = ctx
	request.DocumentVersion++
	request.SnapshotDigest = transcriptDigest("stale")
	result := transcriptAdapter(t, envelope).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Hover != nil {
		t.Fatalf("cancelled result = %#v", result)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != DiagnosticCancelled {
		t.Fatalf("cancelled diagnostics = %#v", result.Diagnostics)
	}
}
