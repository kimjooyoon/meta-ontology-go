package lsp

import (
	"strconv"
	"strings"
)

const hoverProvenanceSchema = "gooo/lsp-hover-provenance/v1"

func hoverSymbolDetail(document document, symbol Symbol) string {
	detail := symbolDetail(symbol)
	provenance := hoverProvenanceDetail(document)
	if provenance == "" {
		return detail
	}
	return detail + "\n\n" + provenance
}

func hoverProvenanceDetail(document document) string {
	if document.cacheKey.sourceDigest == "" ||
		document.result.semanticDigest == "" ||
		document.cacheKey.profileDigest == "" ||
		document.cacheKey.toolchainDigest == "" ||
		document.cacheKey.contractDigest == "" {
		return ""
	}
	value := documentProvenance{
		Schema:          documentProvenanceSchema,
		SourceDigest:    document.cacheKey.sourceDigest,
		SemanticDigest:  document.result.semanticDigest,
		ProfileDigest:   document.cacheKey.profileDigest,
		ToolchainDigest: document.cacheKey.toolchainDigest,
		ContractDigest:  document.cacheKey.contractDigest,
		Symbols:          documentProvenanceSymbols(document.result),
	}
	value.SymbolMapDigest = documentProvenanceSymbolMapDigest(value.Symbols)
	value.SubjectDigest = value.SourceDigest
	value.ProvenanceDigest = documentProvenanceDigest(value)
	return strings.Join([]string{
		"provenance schema: " + hoverProvenanceSchema,
		"source digest: " + value.SourceDigest,
		"semantic digest: " + value.SemanticDigest,
		"profile digest: " + value.ProfileDigest,
		"toolchain digest: " + value.ToolchainDigest,
		"contract digest: " + value.ContractDigest,
		"symbol origin count: " + strconv.Itoa(len(value.Symbols)),
		"symbol map digest: " + value.SymbolMapDigest,
		"provenance digest: " + value.ProvenanceDigest,
	}, "\n")
}
