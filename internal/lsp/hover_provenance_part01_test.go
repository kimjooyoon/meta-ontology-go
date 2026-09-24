package lsp

import (
	"strings"
	"testing"
)

func TestHoverSymbolDetailIncludesExactDocumentProvenance(t *testing.T) {
	document := document{
		cacheKey: documentCacheKey{
			sourceDigest:    "source-digest",
			profileDigest:   "profile-digest",
			toolchainDigest: "toolchain-digest",
			contractDigest:  "contract-digest",
		},
		result: ParseResult{semanticDigest: "semantic-digest"},
	}
	detail := hoverSymbolDetail(document, Symbol{Name: "Order", Detail: "entity Order", ID: "billing://entity/order"})
	expected := documentProvenance{
		Schema:          documentProvenanceSchema,
		SourceDigest:    "source-digest",
		SubjectDigest:   "source-digest",
		SemanticDigest:  "semantic-digest",
		ProfileDigest:   "profile-digest",
		ToolchainDigest: "toolchain-digest",
		ContractDigest:  "contract-digest",
		Symbols:         documentProvenanceSymbols(document.result),
		References:      documentProvenanceReferences(document.result),
	}
	expected.SymbolMapDigest = documentProvenanceSymbolMapDigest(expected.Symbols)
	expected.ReferenceMapDigest = documentProvenanceReferenceMapDigest(expected.References)
	expected.ProvenanceDigest = documentProvenanceDigest(expected)
	for _, fragment := range []string{
		"entity Order (semantic ID: billing://entity/order)",
		"provenance schema: " + hoverProvenanceSchema,
		"source digest: " + expected.SourceDigest,
		"semantic digest: " + expected.SemanticDigest,
		"symbol origin count: 0",
		"symbol map digest: " + expected.SymbolMapDigest,
		"reference origin count: 0",
		"reference map digest: " + expected.ReferenceMapDigest,
		"provenance digest: " + expected.ProvenanceDigest,
	} {
		if !strings.Contains(detail, fragment) {
			t.Fatalf("hover detail missing %q: %q", fragment, detail)
		}
	}
}

func TestHoverSymbolDetailOmitsIncompleteProvenance(t *testing.T) {
	detail := hoverSymbolDetail(document{}, Symbol{Name: "Order", Detail: "entity Order"})
	if detail != "entity Order" {
		t.Fatalf("detail = %q, want existing detail without incomplete provenance", detail)
	}
}
