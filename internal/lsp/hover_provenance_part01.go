package lsp

import (
	"strconv"
	"strings"
)

const hoverProvenanceSchema = "gooo/lsp-hover-provenance/v1"

func hoverSymbolDetail(document document, symbol Symbol) string {
	detail := symbolDetail(symbol)
	provenance := hoverProvenanceDetail(document)
	evidence := hoverSelfImprovementEvidenceDetail(document)
	if provenance == "" && evidence == "" {
		return detail
	}
	sections := []string{detail}
	if provenance != "" {
		sections = append(sections, provenance)
	}
	if evidence != "" {
		sections = append(sections, evidence)
	}
	return strings.Join(sections, "\n\n")
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
		Symbols:         documentProvenanceSymbols(document.result),
		References:      documentProvenanceReferences(document.result),
	}
	value.SymbolMapDigest = documentProvenanceSymbolMapDigest(value.Symbols)
	value.ReferenceMapDigest = documentProvenanceReferenceMapDigest(value.References)
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
		"reference origin count: " + strconv.Itoa(len(value.References)),
		"reference map digest: " + value.ReferenceMapDigest,
		"provenance digest: " + value.ProvenanceDigest,
	}, "\n")
}
