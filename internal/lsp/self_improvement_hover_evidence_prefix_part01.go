package lsp

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

const hoverSelfImprovementEvidenceSchemaPart01 = "gooo/lsp-hover-self-improvement-evidence/v1"

// hoverSelfImprovementEvidenceDetail preserves a deterministic prefix of the
// provenance chain. An incomplete chain is observable as UNKNOWN through its
// first missing stage instead of being presented as a successful hover.
func hoverSelfImprovementEvidenceDetail(document document) string {
	stages := []string{
		document.cacheKey.sourceDigest,
		document.result.semanticDigest,
		document.cacheKey.profileDigest,
		document.cacheKey.toolchainDigest,
		document.cacheKey.contractDigest,
	}
	hasEvidence := false
	for _, stage := range stages {
		if stage != "" {
			hasEvidence = true
			break
		}
	}
	if !hasEvidence {
		return ""
	}
	if stages[0] != "" && stages[1] != "" && stages[2] != "" && stages[3] != "" && stages[4] != "" {
		value := documentProvenance{
			Schema:          documentProvenanceSchema,
			SourceDigest:    stages[0],
			SemanticDigest:  stages[1],
			ProfileDigest:   stages[2],
			ToolchainDigest: stages[3],
			ContractDigest:  stages[4],
			Symbols:         documentProvenanceSymbols(document.result),
			References:      documentProvenanceReferences(document.result),
		}
		value.SymbolMapDigest = documentProvenanceSymbolMapDigest(value.Symbols)
		value.ReferenceMapDigest = documentProvenanceReferenceMapDigest(value.References)
		value.SubjectDigest = value.SourceDigest
		value.ProvenanceDigest = documentProvenanceDigest(value)
		stages = append(stages, value.SymbolMapDigest, value.ReferenceMapDigest, value.ProvenanceDigest)
	}
	missing := len(stages)
	for index, stage := range stages {
		if stage == "" {
			missing = index
			break
		}
	}
	prefix := stages
	missingIndex := -1
	if missing < len(stages) {
		prefix = stages[:missing]
		missingIndex = missing
	}
	return strings.Join([]string{
		"self-improvement evidence schema: " + hoverSelfImprovementEvidenceSchemaPart01,
		"evidence prefix digest: " + hoverSelfImprovementEvidenceDigestPart01(prefix),
		"missing stage index: " + strconv.Itoa(missingIndex),
	}, "\n")
}

func hoverSelfImprovementEvidenceDigestPart01(prefix []string) string {
	payload := hoverSelfImprovementEvidenceSchemaPart01 + "\x00" + strings.Join(prefix, "\x00")
	sum := sha256.Sum256([]byte(payload))
	return "sha256:" + hex.EncodeToString(sum[:])
}
