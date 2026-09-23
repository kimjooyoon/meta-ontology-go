package lsp

import (
	"encoding/json"
	"errors"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

type documentProvenanceReference struct {
	Name         string `json:"name"`
	SemanticID   string `json:"semantic_id"`
	Range        Range  `json:"range"`
	OriginDigest string `json:"origin_digest"`
}

func documentProvenanceReferences(result ParseResult) []documentProvenanceReference {
	values := make([]documentProvenanceReference, 0, len(result.References))
	for _, reference := range result.References {
		value := documentProvenanceReference{Name: reference.Name, SemanticID: reference.ID, Range: reference.Range}
		value.OriginDigest = documentProvenanceReferenceDigest(value)
		values = append(values, value)
	}
	sort.SliceStable(values, func(left, right int) bool {
		first, second := values[left], values[right]
		if first.Name != second.Name {
			return first.Name < second.Name
		}
		if first.SemanticID != second.SemanticID {
			return first.SemanticID < second.SemanticID
		}
		if first.Range.Start != second.Range.Start {
			return positionLess(first.Range.Start, second.Range.Start)
		}
		return positionLess(first.Range.End, second.Range.End)
	})
	return values
}

func documentProvenanceReferenceDigest(value documentProvenanceReference) string {
	value.OriginDigest = ""
	payload, _ := json.Marshal(value)
	return cache.HashBytes(payload).String()
}

func documentProvenanceReferenceMapDigest(values []documentProvenanceReference) string {
	payload, _ := json.Marshal(values)
	return cache.HashBytes(payload).String()
}

func validateDocumentProvenanceReferences(values []documentProvenanceReference) error {
	for _, value := range values {
		if value.Name == "" || !cache.Digest(value.OriginDigest).Known() || value.OriginDigest != documentProvenanceReferenceDigest(value) {
			return errors.New("document provenance reference origin is invalid")
		}
	}
	return nil
}
