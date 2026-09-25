package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestDocumentProvenanceReferenceMapBindsReverseObservation(t *testing.T) {
	first := documentProvenanceReferences(ParseResult{References: []Reference{{
		Name: "Order", ID: "order-id", Range: testRange(0, 0, 0, 5),
	}}})
	second := documentProvenanceReferences(ParseResult{References: []Reference{{
		Name: "Order", ID: "order-id", Range: testRange(0, 1, 0, 6),
	}}})
	firstDigest := documentProvenanceReferenceMapDigest(first)
	secondDigest := documentProvenanceReferenceMapDigest(second)
	if !cache.Digest(firstDigest).Known() || firstDigest == secondDigest {
		t.Fatalf("reference map digest = %q/%q", firstDigest, secondDigest)
	}
	if err := validateDocumentProvenanceReferences(first); err != nil {
		t.Fatalf("validateDocumentProvenanceReferences() error = %v", err)
	}
	tampered := append([]documentProvenanceReference(nil), first...)
	tampered[0].SemanticID = "tampered"
	if err := validateDocumentProvenanceReferences(tampered); err == nil {
		t.Fatal("validateDocumentProvenanceReferences() accepted tampered origin")
	}
}
