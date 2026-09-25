package lsp

import "testing"

func TestDiagnosticProvenanceItemBindsDocumentProvenance(t *testing.T) {
	documentDigest := digestText("package origin\nnamespace origin\n")
	items := diagnosticProvenanceItems([]Diagnostic{{Message: "syntax origin"}}, documentDigest)
	if len(items) != 1 || items[0].DocumentProvenanceDigest != documentDigest {
		t.Fatalf("diagnostic did not retain document provenance: %#v", items)
	}
	if err := validateDiagnosticProvenance(diagnosticProvenanceObservation{
		Schema: diagnosticProvenanceSchema, URI: "file:///origin.gooo", Diagnostics: items,
		DocumentProvenanceDigest: documentDigest, DiagnosticMapDigest: diagnosticProvenanceMapDigest(items),
		Decision: diagnosticProvenanceUnknown, Reason: "MISSING_SOURCE_DIGEST", NonAuthorizing: true,
	}); err == nil {
		t.Fatal("incomplete diagnostic observation was accepted")
	}
}
