package coupling

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
