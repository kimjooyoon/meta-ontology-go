package lsp

import (
	"strings"
	"testing"
)

func TestHoverSelfImprovementEvidenceKeepsFirstMissingStage(t *testing.T) {
	document := document{cacheKey: documentCacheKey{sourceDigest: "source-digest"}}
	detail := hoverSymbolDetail(document, Symbol{Name: "Order", Detail: "entity Order"})
	wantDigest := hoverSelfImprovementEvidenceDigestPart01([]string{"source-digest"})
	for _, fragment := range []string{
		"self-improvement evidence schema: " + hoverSelfImprovementEvidenceSchemaPart01,
		"evidence prefix digest: " + wantDigest,
		"missing stage index: 1",
	} {
		if !strings.Contains(detail, fragment) {
			t.Fatalf("hover detail missing %q: %q", fragment, detail)
		}
	}
}

func TestHoverSelfImprovementEvidenceClosesOnlyWithCompletePrefix(t *testing.T) {
	document := document{
		cacheKey: documentCacheKey{
			sourceDigest:    "source-digest",
			profileDigest:   "profile-digest",
			toolchainDigest: "toolchain-digest",
			contractDigest:  "contract-digest",
		},
		result: ParseResult{semanticDigest: "semantic-digest"},
	}
	detail := hoverSymbolDetail(document, Symbol{Name: "Order", Detail: "entity Order"})
	if !strings.Contains(detail, "missing stage index: -1") {
		t.Fatalf("complete hover evidence was not closed: %q", detail)
	}
}
