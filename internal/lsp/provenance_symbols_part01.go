package lsp

import (
	"encoding/json"
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

type documentProvenanceSymbol struct {
	Name           string     `json:"name"`
	SemanticID     string     `json:"semantic_id"`
	Kind           SymbolKind `json:"kind"`
	Range          Range      `json:"range"`
	SelectionRange Range      `json:"selection_range"`
	OriginDigest   string     `json:"origin_digest"`
}

func documentProvenanceSymbols(result ParseResult) []documentProvenanceSymbol {
	canonical := canonicalDocumentSymbols(allSymbols(result))
	values := make([]documentProvenanceSymbol, 0, len(canonical))
	for _, symbol := range canonical {
		value := documentProvenanceSymbol{
			Name: symbol.Name, SemanticID: symbol.ID, Kind: symbol.Kind,
			Range: symbol.Range, SelectionRange: symbol.SelectionRange,
		}
		value.OriginDigest = documentProvenanceSymbolDigest(value)
		values = append(values, value)
	}
	return values
}

func documentProvenanceSymbolDigest(value documentProvenanceSymbol) string {
	value.OriginDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func documentProvenanceSymbolMapDigest(values []documentProvenanceSymbol) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func validateDocumentProvenanceSymbols(values []documentProvenanceSymbol) error {
	for _, value := range values {
		if value.Name == "" || !cache.Digest(value.OriginDigest).Known() || value.OriginDigest != documentProvenanceSymbolDigest(value) {
			return errors.New("document provenance symbol origin is invalid")
		}
	}
	return nil
}
