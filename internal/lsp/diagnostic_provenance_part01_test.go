package lsp

import (
	"context"
	"encoding/json"
	"testing"
)

func TestDiagnosticProvenanceBindsDocumentAndDiagnostics(t *testing.T) {
	server := NewServer()
	uri := "file:///diagnostics.gooo"
	source := "package diagnostics\nnamespace diagnostics\n"
	server.documents[uri] = &document{text: source}
	params, err := json.Marshal(diagnosticProvenanceParams{TextDocument: TextDocumentIdentifier{URI: uri}})
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := server.diagnosticProvenanceRequest(context.Background(), requestEnvelope{ID: json.RawMessage("1"), Params: params})
	if err != nil {
		t.Fatal(err)
	}
	if response == nil || len(response.Result) == 0 {
		t.Fatalf("missing diagnostic provenance response: %#v", response)
	}
	value, err := decodeDiagnosticProvenance(response.Result)
	if err != nil {
		t.Fatal(err)
	}
	if value.Schema != diagnosticProvenanceSchema || value.URI != uri || value.Decision != diagnosticProvenanceClosed ||
		value.Reason != "DIAGNOSTIC_SURFACE_BOUND" || value.DocumentProvenanceDigest == "" ||
		value.DiagnosticMapDigest == "" || value.ObservationDigest == "" {
		t.Fatalf("incomplete diagnostic provenance: %#v", value)
	}
	if err := validateDiagnosticProvenance(value); err != nil {
		t.Fatal(err)
	}
	changedSource := "package diagnostics\nnamespace changed\n"
	changeParams, err := json.Marshal(map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 2},
		"contentChanges": []any{map[string]any{"text": changedSource}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := server.didChange(context.Background(), requestEnvelope{Params: changeParams}); err != nil {
		t.Fatal(err)
	}
	response, _, err = server.diagnosticProvenanceRequest(context.Background(), requestEnvelope{ID: json.RawMessage("2"), Params: params})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := decodeDiagnosticProvenance(response.Result)
	if err != nil {
		t.Fatal(err)
	}
	if changed.SourceDigest == value.SourceDigest || changed.DocumentProvenanceDigest == value.DocumentProvenanceDigest ||
		changed.ObservationDigest == value.ObservationDigest {
		t.Fatalf("diagnostic provenance was stale after change: before=%#v after=%#v", value, changed)
	}
	tampered := changed
	tampered.Diagnostics = append(tampered.Diagnostics, diagnosticProvenanceItem{Message: "tampered"})
	if validateDiagnosticProvenance(tampered) == nil {
		t.Fatal("tampered diagnostic provenance was accepted")
	}
}

func TestDiagnosticProvenanceUnknownWithoutSourceDigest(t *testing.T) {
	uri := "file:///unknown-diagnostics.gooo"
	value := observeDiagnosticProvenance(uri, document{result: ParseResult{}}, documentCacheKey{})
	if value.Decision != diagnosticProvenanceUnknown || value.Reason != "MISSING_SOURCE_DIGEST" || !value.NonAuthorizing {
		t.Fatalf("unexpected unknown diagnostic provenance: %#v", value)
	}
	if err := validateDiagnosticProvenance(value); err != nil {
		t.Fatal(err)
	}
}
